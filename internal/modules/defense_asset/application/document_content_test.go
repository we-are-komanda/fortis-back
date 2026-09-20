package application

import (
	"bytes"
	"github.com/fortis/backend/internal/modules/defense_asset/domain"
	"github.com/stretchr/testify/require"
	"image"
	"image/jpeg"
	"image/png"
	"testing"
)

func TestDocumentContentAllowlist(t *testing.T) {
	var pngBody, jpegBody bytes.Buffer
	require.NoError(t, png.Encode(&pngBody, image.NewRGBA(image.Rect(0, 0, 2, 2))))
	require.NoError(t, jpeg.Encode(&jpegBody, image.NewRGBA(image.Rect(0, 0, 2, 2)), nil))
	for _, in := range []UploadDocumentInput{{Name: "../source.txt", MimeType: "text/plain", Body: []byte("Synthetic source")}, {Name: "source.pdf", MimeType: "application/pdf", Body: []byte("%PDF-1.4\nsynthetic fixture\n%%EOF")}, {Name: "source.png", MimeType: "image/png", Body: pngBody.Bytes()}, {Name: "source.jpg", MimeType: "image/jpeg", Body: jpegBody.Bytes()}} {
		name, kind, err := validatedDocument(in)
		require.NoError(t, err)
		require.NotContains(t, name, "..")
		require.Equal(t, in.MimeType, kind)
	}
	for _, in := range []UploadDocumentInput{{Name: "payload.txt", MimeType: "text/plain", Body: []byte("<html><script>example</script></html>")}, {Name: "payload.pdf", MimeType: "application/pdf", Body: []byte("not PDF")}, {Name: "payload.svg", MimeType: "image/svg+xml", Body: []byte("<svg/>")}, {Name: "empty.txt", MimeType: "text/plain", Body: nil}} {
		_, _, err := validatedDocument(in)
		require.ErrorIs(t, err, domain.ErrDocumentMediaType)
	}
}
