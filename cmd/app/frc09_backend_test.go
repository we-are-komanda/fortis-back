package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/fortis/backend/internal/auth"
	aa "github.com/fortis/backend/internal/modules/defense_asset/application"
	ad "github.com/fortis/backend/internal/modules/defense_asset/domain"
	ai "github.com/fortis/backend/internal/modules/defense_asset/infrastructure"
	au "github.com/fortis/backend/internal/modules/defense_asset/ui"
	ea "github.com/fortis/backend/internal/modules/enterprise/application"
	ei "github.com/fortis/backend/internal/modules/enterprise/infrastructure"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"
	"gorm.io/gorm"
	"mime/multipart"
	"net/textproto"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestFRC09HTTPUploadAndDownload(t *testing.T) {
	_, f, s, _, scanner := frc09Services(t)
	a, b := f.people[0], f.people[1]
	controller := au.NewDocumentController(s)
	upload := func(actor, asset, kind, name string, body []byte) *fasthttp.Response {
		var buffer bytes.Buffer
		w := multipart.NewWriter(&buffer)
		require.NoError(t, w.WriteField("assetId", asset))
		header := textproto.MIMEHeader{}
		header.Set("Content-Disposition", `form-data; name="file"; filename="`+name+`"`)
		header.Set("Content-Type", kind)
		part, err := w.CreatePart(header)
		require.NoError(t, err)
		_, err = part.Write(body)
		require.NoError(t, err)
		require.NoError(t, w.Close())
		var ctx fasthttp.RequestCtx
		ctx.Init2(nil, nil, false)
		ctx.Request.Header.SetMethod("POST")
		ctx.Request.Header.SetContentType(w.FormDataContentType())
		ctx.Request.SetBody(buffer.Bytes())
		ctx.SetUserValue("userID", actor)
		controller.Create(&ctx)
		response := new(fasthttp.Response)
		ctx.Response.CopyTo(response)
		return response
	}
	response := upload(a.u, a.a, "text/plain", "synthetic.txt", []byte("Synthetic source"))
	require.Equal(t, 201, response.StatusCode(), string(response.Body()))
	var result map[string]any
	require.NoError(t, json.Unmarshal(response.Body(), &result))
	require.Equal(t, true, result["commercial"])
	require.Equal(t, "ready", result["status"])
	require.True(t, strings.HasPrefix(result["downloadUrl"].(string), "/api/v1/assets/documents/download?"))
	var download fasthttp.RequestCtx
	download.Init2(nil, nil, false)
	download.SetUserValue("userID", a.u)
	download.QueryArgs().Set("id", result["id"].(string))
	controller.Download(&download)
	require.Equal(t, 200, download.Response.StatusCode())
	require.Equal(t, "Synthetic source", string(download.Response.Body()))
	require.Equal(t, "nosniff", string(download.Response.Header.Peek("X-Content-Type-Options")))
	require.Contains(t, string(download.Response.Header.Peek("Content-Disposition")), "attachment")
	require.Equal(t, "private, no-store", string(download.Response.Header.Peek("Cache-Control")))
	download.SetUserValue("userID", b.u)
	controller.Download(&download)
	require.Equal(t, 404, download.Response.StatusCode())
	require.Equal(t, 404, upload(b.u, a.a, "text/plain", "synthetic.txt", []byte("Synthetic")).StatusCode())
	require.Equal(t, 403, upload(a.u, f.publicAsset, "text/plain", "synthetic.txt", []byte("Synthetic")).StatusCode())
	require.Equal(t, 415, upload(a.u, a.a, "application/pdf", "synthetic.pdf", []byte("Not PDF")).StatusCode())
	require.Equal(t, 413, upload(a.u, a.a, "text/plain", "large.txt", bytes.Repeat([]byte("a"), ad.MaxDocumentBytes+1)).StatusCode())
	scanner.err = ad.ErrDocumentUnavailable
	require.Equal(t, 503, upload(a.u, a.a, "text/plain", "quarantine.txt", []byte("Synthetic")).StatusCode())
	scanner.err = ad.ErrDocumentRejected
	require.Equal(t, 422, upload(a.u, a.a, "text/plain", "rejected.txt", []byte("Synthetic")).StatusCode())
	var legacy fasthttp.RequestCtx
	legacy.Init2(nil, nil, false)
	legacy.SetUserValue("userID", a.u)
	legacy.Request.Header.SetContentType("application/json")
	legacy.Request.SetBodyString(`{"assetId":"` + a.a + `","name":"remote.txt","storageKey":"untrusted","downloadUrl":"https://example.test/public"}`)
	controller.Create(&legacy)
	require.Equal(t, 400, legacy.Response.StatusCode())
	require.Contains(t, string(legacy.Response.Body()), "document_upload_required")
}

type frc09Scanner struct {
	err   error
	calls int
	run   func()
}

func (s *frc09Scanner) Ready() bool { return true }
func (s *frc09Scanner) Scan(context.Context, string) error {
	s.calls++
	if s.run != nil {
		s.run()
	}
	return s.err
}
func frc09Services(t *testing.T) (*gorm.DB, accessFixture, *aa.DocumentService, *aa.DefenseAssetService, *frc09Scanner) {
	t.Helper()
	dsn := os.Getenv("FRC09_TEST_DSN")
	if dsn == "" {
		t.Skip("FRC09_TEST_DSN unset")
	}
	t.Setenv("FORTIS_FRC04_TEST_DSN", dsn)
	db := frc04Database(t)
	f := newPostgresFixture(t, db)
	repo := ai.NewDocumentRepository(db)
	assets := aa.NewDefenseAssetService(ai.NewDefenseAssetRepository(db), ea.NewEnterpriseService(ei.NewEnterpriseRepository(db)), repo)
	scanner := &frc09Scanner{}
	storage := ai.NewLocalDocumentStorage(ai.DocumentStorageConfig{RootDir: filepath.Join(t.TempDir(), "private"), Enabled: true})
	return db, f, aa.NewDocumentService(repo, assets, aa.DocumentPipeline{Storage: storage, Scanner: scanner}), assets, scanner
}
func TestFRC09DocumentsPrivateImmutableAndProvenance(t *testing.T) {
	db, f, s, assets, _ := frc09Services(t)
	a, b := f.people[0], f.people[1]
	ctx := context.Background()
	in := aa.UploadDocumentInput{AssetID: a.a, Name: "../synthetic.txt", MimeType: "text/plain", Body: []byte("synthetic source"), Commercial: true}
	doc, err := s.Upload(ctx, a.u, in)
	require.NoError(t, err)
	require.Equal(t, "ready", doc.Status())
	require.Equal(t, "synthetic.txt", doc.Name())
	require.Len(t, doc.Checksum(), 64)
	_, body, err := s.Download(ctx, a.u, doc.ID(), a.a)
	require.NoError(t, err)
	require.Equal(t, in.Body, body)
	_, _, err = s.Download(ctx, b.u, doc.ID(), "")
	require.ErrorIs(t, err, auth.ErrNotFound)
	_, _, err = s.Download(ctx, a.u, doc.ID(), b.a)
	require.ErrorIs(t, err, auth.ErrNotFound)
	newer, err := s.Upload(ctx, a.u, in)
	require.NoError(t, err)
	require.NotEqual(t, doc.ID(), newer.ID())
	require.Error(t, db.Table("defense_asset_documents").Where("id = ?", doc.ID()).Update("checksum", strings.Repeat("0", 64)).Error)
	date := "2026-09-20"
	id := doc.ID()
	asset, err := assets.Update(ctx, a.u, aa.UpdateInput{ID: a.a, CatalogMetadataInput: aa.CatalogMetadataInput{Provenance: &aa.ProvenanceInput{SourceDocumentID: &id, SourceDate: &date, Quality: "confirmed"}}})
	require.NoError(t, err)
	require.Equal(t, a.u, asset.Provenance().RecordedBy())
	require.Equal(t, doc.Checksum(), asset.Provenance().DocumentChecksum())
	require.NotEmpty(t, asset.Provenance().Revision())
	require.NoError(t, s.Delete(ctx, a.u, id))
	_, _, err = s.Download(ctx, a.u, id, "")
	require.Error(t, err)
	require.Equal(t, doc.Checksum(), asset.Provenance().DocumentChecksum())
	var count int64
	require.NoError(t, db.Table("audit_events").Where("entity_id = ?", id).Count(&count).Error)
	require.EqualValues(t, 3, count)
}
func TestFRC09QuarantineValidationAndAtomicity(t *testing.T) {
	db, f, s, _, scanner := frc09Services(t)
	a := f.people[0]
	ctx := context.Background()
	in := aa.UploadDocumentInput{AssetID: a.a, Name: "sample.txt", MimeType: "text/plain", Body: []byte("synthetic text")}
	scanner.err = ad.ErrDocumentUnavailable
	doc, err := s.Upload(ctx, a.u, in)
	require.ErrorIs(t, err, ad.ErrDocumentUnavailable)
	require.NotNil(t, doc)
	require.Equal(t, "quarantined", doc.Status())
	_, _, err = s.Download(ctx, a.u, doc.ID(), "")
	require.ErrorIs(t, err, ad.ErrDocumentNotReady)
	scanner.err = ad.ErrDocumentRejected
	doc, err = s.Upload(ctx, a.u, in)
	require.ErrorIs(t, err, ad.ErrDocumentRejected)
	require.Equal(t, "rejected", doc.Status())
	scanner.err = nil
	bad := in
	bad.MimeType = "application/pdf"
	_, err = s.Upload(ctx, a.u, bad)
	require.ErrorIs(t, err, ad.ErrDocumentMediaType)
	bad = in
	bad.Body = make([]byte, ad.MaxDocumentBytes+1)
	_, err = s.Upload(ctx, a.u, bad)
	require.ErrorIs(t, err, ad.ErrDocumentTooLarge)
	var before, after int64
	require.NoError(t, db.Table("defense_asset_documents").Count(&before).Error)
	require.NoError(t, db.Callback().Create().Before("gorm:create").Register("frc09_audit_failure", func(tx *gorm.DB) {
		if tx.Statement.Table == "audit_events" {
			tx.AddError(errors.New("synthetic audit failure"))
		}
	}))
	_, err = s.Upload(ctx, a.u, in)
	require.Error(t, err)
	require.NoError(t, db.Callback().Create().Remove("frc09_audit_failure"))
	require.NoError(t, db.Table("defense_asset_documents").Count(&after).Error)
	require.Equal(t, before, after)
	legacy, err := ad.NewDocument("00000000-0000-4000-8000-000000000909", a.a, "legacy", "text/plain", "untrusted", "https://example.test/public", 1, &a.u, time.Now(), time.Now())
	require.NoError(t, err)
	require.NoError(t, ai.NewDocumentRepository(db).Save(ctx, legacy))
	_, _, err = s.Download(ctx, a.u, legacy.ID(), "")
	require.ErrorIs(t, err, ad.ErrDocumentNotReady)
}

func TestFRC09CatalogMoneyTypedCardsAndProvenance(t *testing.T) {
	db, f, _, assets, _ := frc09Services(t)
	a := f.people[0]
	controller := au.NewDefenseAssetController(assets)
	roles := []string{"detect", "track", "classify", "suppress", "destroy", "delay", "protect", "coordinate", "monitor", "alert"}
	categories := []string{"early-warning", "detection", "classification", "jamming", "spoofing", "kinetic", "interceptor", "passive-protection", "engineering-protection", "infrastructure", "software", "command-center", "external-service"}
	for _, category := range categories {
		t.Run(category, func(t *testing.T) {
			body := map[string]any{"name": "Synthetic card", "category": category, "roles": roles, "coverageType": "none", "enterpriseId": a.e, "pricePerUnitMln": 999, "unitPriceMinor": nil, "pricingMode": "components", "components": []any{map[string]any{"id": "part", "name": "Synthetic component", "quantity": 2, "unitPriceMinor": "125"}}, "weaponSpec": map[string]any{"caliber": "synthetic", "ammunitionType": "synthetic", "operationMode": "fixture", "moduleCount": 1, "isManual": false}, "detectionSpec": map[string]any{"frequencyRange": "synthetic", "detectionMode": "fixture", "rotationSpeed": 0, "fieldOfView": 0, "hasThermalImager": false}, "ewSpec": map[string]any{"frequencyRange": "synthetic", "actionRange": 0, "azimuth": 0}, "provenance": map[string]any{"sourceLabel": "Synthetic data", "quality": "demo", "recordedBy": "forged", "recordedAt": "forged", "revision": "forged"}}
			switch category {
			case "early-warning", "detection", "classification":
				delete(body, "weaponSpec")
				delete(body, "ewSpec")
			case "jamming", "spoofing":
				delete(body, "weaponSpec")
				delete(body, "detectionSpec")
			case "kinetic", "interceptor":
				delete(body, "detectionSpec")
				delete(body, "ewSpec")
			default:
				delete(body, "weaponSpec")
				delete(body, "detectionSpec")
				delete(body, "ewSpec")
			}
			var ctx fasthttp.RequestCtx
			ctx.Init2(nil, nil, false)
			ctx.SetUserValue("userID", a.u)
			ctx.Request.SetBodyString(js(body))
			controller.Create(&ctx)
			require.Equal(t, 201, ctx.Response.StatusCode(), string(ctx.Response.Body()))
			var created map[string]any
			require.NoError(t, json.Unmarshal(ctx.Response.Body(), &created))
			id := created["id"].(string)
			loaded, err := ai.NewDefenseAssetRepository(db).FindByID(context.Background(), id)
			require.NoError(t, err)
			require.True(t, loaded.HasUnitPriceMinor())
			require.Nil(t, loaded.UnitPriceMinor())
			require.Equal(t, category, string(loaded.Category()))
			require.Len(t, loaded.Roles(), len(roles))
			if body["weaponSpec"] != nil {
				require.NotNil(t, loaded.WeaponSpec())
				require.Equal(t, "fixture", *loaded.WeaponSpec().OperationMode)
			}
			if body["detectionSpec"] != nil {
				require.NotNil(t, loaded.DetectionSpec())
				require.False(t, *loaded.DetectionSpec().HasThermalImager)
			}
			if body["ewSpec"] != nil {
				require.NotNil(t, loaded.EWSpec())
				require.Zero(t, *loaded.EWSpec().ActionRange)
			}
			require.Equal(t, a.u, loaded.Provenance().RecordedBy())
			require.NotEqual(t, "forged", loaded.Provenance().RecordedAt())
			require.NotEqual(t, "forged", loaded.Provenance().Revision())
			require.Equal(t, "demo", loaded.Provenance().Quality())
			for _, key := range []string{"category", "roles", "unitPriceMinor", "pricingMode", "components", "weaponSpec", "detectionSpec", "ewSpec"} {
				require.JSONEq(t, js(body[key]), js(created[key]), key)
			}
		})
	}
	var missing fasthttp.RequestCtx
	missing.Init2(nil, nil, false)
	missing.SetUserValue("userID", a.u)
	missing.Request.SetBodyString(js(map[string]any{"name": "Invalid confirmed", "category": "detection", "coverageType": "none", "enterpriseId": a.e, "provenance": map[string]any{"quality": "confirmed", "sourceLabel": ""}}))
	controller.Create(&missing)
	require.Equal(t, 400, missing.Response.StatusCode())
}

func TestFRC09RevocationDuringScanPreventsPublication(t *testing.T) {
	db, f, s, assets, scanner := frc09Services(t)
	a := f.people[0]
	ctx := context.Background()
	closed := aa.NewDocumentService(ai.NewDocumentRepository(db), assets)
	_, err := closed.Upload(ctx, a.u, aa.UploadDocumentInput{AssetID: a.a, Name: "closed.txt", MimeType: "text/plain", Body: []byte("Synthetic")})
	require.ErrorIs(t, err, ad.ErrDocumentUnavailable)
	scanner.run = func() { require.NoError(t, ei.NewEnterpriseRepository(db).RemoveUserFromEnterprise(ctx, a.u, a.e)) }
	_, err = s.Upload(ctx, a.u, aa.UploadDocumentInput{AssetID: a.a, Name: "revoked.txt", MimeType: "text/plain", Body: []byte("Synthetic")})
	require.ErrorIs(t, err, auth.ErrNotFound)
	var rows []ai.DefenseAssetDocumentModel
	require.NoError(t, db.Where("name = ?", "revoked.txt").Find(&rows).Error)
	require.Len(t, rows, 1)
	require.Equal(t, "quarantined", rows[0].Status)
	var count int64
	require.NoError(t, db.Table("audit_events").Where("entity_id = ? AND action = 'document.ready'", rows[0].ID).Count(&count).Error)
	require.Zero(t, count)
}

func TestFRC09CatalogRejectedUpdatePreservesStoredCard(t *testing.T) {
	db, f, _, assets, _ := frc09Services(t)
	a := f.people[0]
	created, err := assets.Create(context.Background(), a.u, aa.CreateInput{EnterpriseID: &a.e, Name: "Original", Category: "detection", CoverageType: "none", DetectionSpec: &ad.DetectionSpecification{}})
	require.NoError(t, err)
	controller := au.NewDefenseAssetController(assets)
	for _, change := range []string{`{"name":"Rejected","category":"kinetic"}`, `{"name":"Rejected","weaponSpec":{}}`, `{"name":"Rejected","deploymentType":"invalid"}`} {
		var ctx fasthttp.RequestCtx
		ctx.Init2(nil, nil, false)
		ctx.SetUserValue("userID", a.u)
		ctx.QueryArgs().Set("id", created.ID())
		ctx.Request.SetBodyString(change)
		controller.Update(&ctx)
		require.Equal(t, 400, ctx.Response.StatusCode(), string(ctx.Response.Body()))
		loaded, err := ai.NewDefenseAssetRepository(db).FindByID(context.Background(), created.ID())
		require.NoError(t, err)
		require.Equal(t, "Original", loaded.Name())
		require.Equal(t, ad.DefenseAssetCategory("detection"), loaded.Category())
		require.NotNil(t, loaded.DetectionSpec())
		require.Nil(t, loaded.WeaponSpec())
	}
}
