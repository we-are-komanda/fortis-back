-- Relocate only the unchanged synthetic starting site; custom enterprises are untouched.
UPDATE enterprises SET latitude = -41.2865, longitude = 174.7762,
    address = 'Новая Зеландия, Веллингтон — условная площадка', updated_at = NOW()
WHERE id = 'f0711500-0000-4000-8000-000000000001'
    AND latitude = 55.0 AND longitude = 40.0;
