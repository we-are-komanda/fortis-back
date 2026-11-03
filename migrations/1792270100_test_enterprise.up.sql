-- Single synthetic starting site. Existing enterprises and projects are untouched.
INSERT INTO enterprises (id, name, address, status, latitude, longitude)
VALUES ('f0711500-0000-4000-8000-000000000001', 'Тестовое предприятие', 'Условная учебная площадка', 'active', 55.0, 40.0)
ON CONFLICT (id) DO NOTHING;
