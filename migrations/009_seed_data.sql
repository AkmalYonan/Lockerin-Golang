-- Migration: 009_seed_data.sql
-- Description: Comprehensive realistic dummy/seed data for Lockerin Smart Locker ecosystem

-- 1. Profiles (Users, Admins, Operators)
INSERT INTO profiles (id, email, username, name, phone, avatar_url, role, status, created_at, updated_at)
VALUES 
    ('11111111-1111-1111-1111-111111111111', 'admin@lockerin.id', 'admin', 'Super Admin Lockerin', '081299998888', 'https://images.unsplash.com/photo-1534528741775-53994a69daeb?w=200', 'admin', 'active', NOW() - INTERVAL '30 days', NOW()),
    ('22222222-2222-2222-2222-222222222222', 'akmal@lockerin.id', 'akmal', 'Akmal Maindata', '081234567890', 'https://images.unsplash.com/photo-1507003211169-0a1dd7228f2d?w=200', 'user', 'active', NOW() - INTERVAL '20 days', NOW()),
    ('33333333-3333-3333-3333-333333333333', 'operator@lockerin.id', 'operator', 'Field Operator Jakarta', '081288887777', 'https://images.unsplash.com/photo-1500648767791-00dcc994a43e?w=200', 'operator', 'active', NOW() - INTERVAL '15 days', NOW()),
    ('44444444-4444-4444-4444-444444444444', 'budi.santoso@gmail.com', 'budisantoso', 'Budi Santoso', '081311223344', 'https://images.unsplash.com/photo-1492562080023-ab3db95bfbce?w=200', 'user', 'active', NOW() - INTERVAL '10 days', NOW()),
    ('55555555-5555-5555-5555-555555555555', 'siti.rahma@gmail.com', 'sitirahma', 'Siti Rahmawati', '081399887766', 'https://images.unsplash.com/photo-1544005313-94ddf0286df2?w=200', 'user', 'active', NOW() - INTERVAL '5 days', NOW())
ON CONFLICT (email) DO UPDATE 
SET name = EXCLUDED.name, phone = EXCLUDED.phone, role = EXCLUDED.role, status = EXCLUDED.status;

-- 2. Locations (Stasiun Loker Strategis di Jakarta)
INSERT INTO locations (id, code, name, address, city, latitude, longitude, status, operating_hours, image_url, created_at, updated_at)
VALUES 
    ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'LOC-MEDISTRA', 'Lockerin RS Medistra', 'Jl. Jend. Gatot Subroto No.Kav. 59, Jakarta Selatan', 'Jakarta Selatan', -6.237200, 106.833500, 'active', '{"open": "06:00", "close": "23:00", "is_24_hours": false}'::jsonb, 'https://images.unsplash.com/photo-1590381105924-c72589b9ef3f?w=800', NOW() - INTERVAL '30 days', NOW()),
    ('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 'LOC-PARAMADINA', 'Lockerin Univ. Paramadina', 'Jl. Gatot Subroto No.Kav. 97, Mampang Prapatan', 'Jakarta Selatan', -6.241500, 106.832900, 'active', '{"open": "07:00", "close": "22:00", "is_24_hours": false}'::jsonb, 'https://images.unsplash.com/photo-1541829070764-84a7d30dd3f3?w=800', NOW() - INTERVAL '30 days', NOW()),
    ('cccccccc-cccc-cccc-cccc-cccccccccccc', 'LOC-KUNINGAN', 'Lockerin Mall Kuningan City', 'Jl. Prof. DR. Satrio No.18, Kuningan, Karet Kuningan', 'Jakarta Selatan', -6.224500, 106.829800, 'active', '{"open": "10:00", "close": "22:00", "is_24_hours": false}'::jsonb, 'https://images.unsplash.com/photo-1555396273-367ea4eb4db5?w=800', NOW() - INTERVAL '25 days', NOW()),
    ('dddddddd-dddd-dddd-dddd-dddddddddddd', 'LOC-TEBET', 'Lockerin Stasiun Tebet', 'Jl. Lapangan Roos Raya, Tebet Timur', 'Jakarta Selatan', -6.226300, 106.858200, 'active', '{"open": "05:00", "close": "24:00", "is_24_hours": false}'::jsonb, 'https://images.unsplash.com/photo-1568605117036-5fe5e7bab0b7?w=800', NOW() - INTERVAL '20 days', NOW()),
    ('eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee', 'LOC-GI', 'Lockerin Grand Indonesia Mall', 'Jl. M.H. Thamrin No.1, Menteng, Jakarta Pusat', 'Jakarta Pusat', -6.195200, 106.820200, 'active', '{"open": "10:00", "close": "22:00", "is_24_hours": false}'::jsonb, 'https://images.unsplash.com/photo-1519642918688-7e43b19245d8?w=800', NOW() - INTERVAL '15 days', NOW())
ON CONFLICT (code) DO UPDATE 
SET name = EXCLUDED.name, address = EXCLUDED.address, latitude = EXCLUDED.latitude, longitude = EXCLUDED.longitude, operating_hours = EXCLUDED.operating_hours, image_url = EXCLUDED.image_url;

-- 3. Lockers (Hardware Units)
INSERT INTO lockers (id, location_id, code, name, device_id, status, firmware_version, last_seen_at, created_at, updated_at)
VALUES 
    ('c1111111-1111-1111-1111-111111111111', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'LCK-MEDISTRA-01', 'Lockerin Medistra Unit 1 (Lantai 1)', 'HEAD-LOCKER-MDS-001', 'online', '2.1.0', NOW(), NOW() - INTERVAL '30 days', NOW()),
    ('c2222222-2222-2222-2222-222222222222', 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 'LCK-PARAMADINA-01', 'Lockerin Paramadina Unit 1 (Gedung A)', 'HEAD-LOCKER-PRM-001', 'online', '2.1.0', NOW(), NOW() - INTERVAL '30 days', NOW()),
    ('c3333333-3333-3333-3333-333333333333', 'cccccccc-cccc-cccc-cccc-cccccccccccc', 'LCK-KUNINGAN-01', 'Lockerin Kuningan City Unit 1 (Lobby LG)', 'HEAD-LOCKER-KNG-001', 'online', '2.0.4', NOW(), NOW() - INTERVAL '25 days', NOW()),
    ('c4444444-4444-4444-4444-444444444444', 'dddddddd-dddd-dddd-dddd-dddddddddddd', 'LCK-TEBET-01', 'Lockerin Stasiun Tebet Unit 1 (Pintu Barat)', 'HEAD-LOCKER-TBT-001', 'online', '2.1.0', NOW(), NOW() - INTERVAL '20 days', NOW()),
    ('c5555555-5555-5555-5555-555555555555', 'eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee', 'LCK-GI-01', 'Lockerin Grand Indonesia Unit 1 (West Mall B1)', 'HEAD-LOCKER-GI-001', 'online', '2.1.0', NOW(), NOW() - INTERVAL '15 days', NOW())
ON CONFLICT (code) DO UPDATE 
SET name = EXCLUDED.name, device_id = EXCLUDED.device_id, status = EXCLUDED.status, firmware_version = EXCLUDED.firmware_version, last_seen_at = NOW();

-- 4. Locker Slots (A1 to E5 for each Locker Unit - 125 Slots total)
DO $$
DECLARE
    loc_record RECORD;
    row_char TEXT;
    col_num INT;
    slot_c TEXT;
    sz TEXT;
    prc NUMERIC(12,2);
    init_status TEXT;
BEGIN
    FOR loc_record IN SELECT id, code FROM lockers LOOP
        FOREACH row_char IN ARRAY ARRAY['A', 'B', 'C', 'D', 'E'] LOOP
            FOR col_num IN 1..5 LOOP
                slot_c := row_char || col_num;
                
                -- Determine Size & Price
                IF col_num = 1 THEN
                    sz := 'Small';
                    prc := 4000.00;
                ELSIF col_num = 5 THEN
                    sz := 'Large';
                    prc := 9000.00;
                ELSE
                    sz := 'Medium';
                    prc := 6000.00;
                END IF;

                -- Default Status
                init_status := 'available';
                IF (row_char = 'A' AND col_num = 2) OR (row_char = 'C' AND col_num = 4) THEN
                    init_status := 'occupied';
                ELSIF (row_char = 'E' AND col_num = 5) THEN
                    init_status := 'maintenance';
                END IF;

                INSERT INTO locker_slots (id, locker_id, slot_code, size, status, base_price_per_hour, created_at, updated_at)
                VALUES (gen_random_uuid(), loc_record.id, slot_c, sz, init_status, prc, NOW() - INTERVAL '30 days', NOW())
                ON CONFLICT (locker_id, slot_code) DO NOTHING;
            END LOOP;
        END LOOP;
    END LOOP;
END $$;

-- 5. Promos / Voucher Diskon
INSERT INTO promos (id, code, title, description, discount_type, discount_value, min_order_amount, max_discount_amount, starts_at, ends_at, status, created_at, updated_at)
VALUES 
    ('f1111111-1111-1111-1111-111111111111', 'LOCKERIN50', 'Diskon Pengguna Baru 50%', 'Diskon 50% hingga Rp 10.000 untuk pengguna baru Lockerin', 'percentage', 50.00, 5000.00, 10000.00, NOW() - INTERVAL '30 days', NOW() + INTERVAL '60 days', 'active', NOW(), NOW()),
    ('f2222222-2222-2222-2222-222222222222', 'WEEKENDHEMAT', 'Promo Weekend Seru 20%', 'Potongan 20% tanpa batas maksimal di akhir pekan', 'percentage', 20.00, 10000.00, 20000.00, NOW() - INTERVAL '10 days', NOW() + INTERVAL '30 days', 'active', NOW(), NOW()),
    ('f3333333-3333-3333-3333-333333333333', 'CASHBACK5RB', 'Potongan Langsung Rp 5.000', 'Potongan Rp 5.000 untuk minimal sewa 3 jam', 'fixed', 5000.00, 12000.00, 5000.00, NOW() - INTERVAL '5 days', NOW() + INTERVAL '20 days', 'active', NOW(), NOW())
ON CONFLICT (code) DO UPDATE 
SET title = EXCLUDED.title, description = EXCLUDED.description, discount_value = EXCLUDED.discount_value, status = EXCLUDED.status;

-- 6. Sample Rentals & Transactions
DO $$
DECLARE
    user_akmal UUID;
    user_budi  UUID;
    user_admin UUID;
    loc_medistra UUID := 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa';
    lck_medistra UUID := 'c1111111-1111-1111-1111-111111111111';
    slot_b3 UUID;
    slot_a2 UUID;
    rent1_id UUID := '77777777-7777-7777-7777-777777777771';
    rent2_id UUID := '77777777-7777-7777-7777-777777777772';
BEGIN
    SELECT id INTO user_akmal FROM profiles WHERE email = 'akmal@lockerin.id' LIMIT 1;
    SELECT id INTO user_budi  FROM profiles WHERE email = 'budi.santoso@gmail.com' LIMIT 1;
    SELECT id INTO user_admin FROM profiles WHERE email = 'admin@lockerin.id' LIMIT 1;

    SELECT id INTO slot_b3 FROM locker_slots WHERE locker_id = lck_medistra AND slot_code = 'B3' LIMIT 1;
    SELECT id INTO slot_a2 FROM locker_slots WHERE locker_id = lck_medistra AND slot_code = 'A2' LIMIT 1;

    -- Rental 1: Active Rental (Akmal at B3)
    IF user_akmal IS NOT NULL AND slot_b3 IS NOT NULL THEN
        INSERT INTO rentals (id, user_id, location_id, locker_id, slot_id, status, duration_hours, base_price_snapshot, discount_amount, total_amount, started_at, expires_at, created_at, updated_at)
        VALUES (
            rent1_id, user_akmal, loc_medistra, lck_medistra, slot_b3, 'active', 4, 6000.00, 5000.00, 19000.00,
            NOW() - INTERVAL '1 hour', NOW() + INTERVAL '3 hours', NOW() - INTERVAL '1 hour', NOW()
        ) ON CONFLICT (id) DO NOTHING;

        -- Update slot current rental
        UPDATE locker_slots SET status = 'occupied', current_rental_id = rent1_id WHERE id = slot_b3;

        -- Security PIN (Hash for PIN '123456' with salt 'salthash123')
        INSERT INTO security_codes (id, rental_id, pin_hash, salt, attempts, max_attempts, expires_at, created_at, updated_at)
        VALUES (
            gen_random_uuid(), rent1_id,
            '95a7090886ff010cf4f7943c5b8b9826042dbbbdfac1551a0279d0ec094f31c2', 'salthash123',
            0, 5, NOW() + INTERVAL '24 hours', NOW() - INTERVAL '1 hour', NOW()
        ) ON CONFLICT (rental_id) DO NOTHING;

        -- Payment for Rental 1
        INSERT INTO payments (id, rental_id, provider, order_id, provider_transaction_id, amount, status, payment_type, snap_token, paid_at, created_at, updated_at)
        VALUES (
            gen_random_uuid(), rent1_id, 'midtrans', 'LOCKERIN-MDS-7771', 'MDT-TRX-9892182', 19000.00, 'settlement', 'qris', 'mock-snap-token-7771',
            NOW() - INTERVAL '1 hour', NOW() - INTERVAL '1 hour', NOW()
        ) ON CONFLICT (order_id) DO NOTHING;
    END IF;

    -- Rental 2: Completed Rental (Budi at A2)
    IF user_budi IS NOT NULL AND slot_a2 IS NOT NULL THEN
        INSERT INTO rentals (id, user_id, location_id, locker_id, slot_id, status, duration_hours, base_price_snapshot, discount_amount, total_amount, started_at, expires_at, ended_at, created_at, updated_at)
        VALUES (
            rent2_id, user_budi, loc_medistra, lck_medistra, slot_a2, 'completed', 2, 6000.00, 0.00, 12000.00,
            NOW() - INTERVAL '5 hours', NOW() - INTERVAL '3 hours', NOW() - INTERVAL '3 hours', NOW() - INTERVAL '5 hours', NOW()
        ) ON CONFLICT (id) DO NOTHING;

        -- Payment for Rental 2
        INSERT INTO payments (id, rental_id, provider, order_id, provider_transaction_id, amount, status, payment_type, snap_token, paid_at, created_at, updated_at)
        VALUES (
            gen_random_uuid(), rent2_id, 'midtrans', 'LOCKERIN-MDS-7772', 'MDT-TRX-9892110', 12000.00, 'settlement', 'gopay', 'mock-snap-token-7772',
            NOW() - INTERVAL '5 hours', NOW() - INTERVAL '5 hours', NOW()
        ) ON CONFLICT (order_id) DO NOTHING;
    END IF;

    -- Notifications
    IF user_akmal IS NOT NULL THEN
        INSERT INTO notifications (id, user_id, type, title, body, created_at)
        VALUES 
            (gen_random_uuid(), user_akmal, 'promo', 'Diskon Spesial Pengguna Baru!', 'Gunakan kode promo LOCKERIN50 untuk potongan 50% pemesanan pertama Anda!', NOW() - INTERVAL '2 days'),
            (gen_random_uuid(), user_akmal, 'payment', 'Pembayaran Loker Berhasil', 'Sewa loker B3 di RS Medistra berhasil dibayar. PIN Keamanan Anda adalah 123456.', NOW() - INTERVAL '1 hour');
    END IF;

    IF user_budi IS NOT NULL THEN
        INSERT INTO notifications (id, user_id, type, title, body, created_at)
        VALUES 
            (gen_random_uuid(), user_budi, 'rental', 'Sewa Selesai', 'Terima kasih telah menggunakan Lockerin di RS Medistra. Sampai jumpa kembali!', NOW() - INTERVAL '3 hours');
    END IF;

    -- Audit Logs
    IF user_akmal IS NOT NULL THEN
        INSERT INTO audit_logs (id, actor_type, actor_id, action, entity_type, entity_id, ip_address, user_agent, created_at)
        VALUES 
            (gen_random_uuid(), 'user', user_akmal::text, 'RESERVE_SLOT', 'rental', '77777777-7777-7777-7777-777777777771', '127.0.0.1', 'Lockerin Flutter App / Android 14', NOW() - INTERVAL '1 hour 5 minutes'),
            (gen_random_uuid(), 'system', 'midtrans_webhook', 'PAYMENT_SETTLEMENT', 'payment', 'LOCKERIN-MDS-7771', '103.10.20.30', 'Midtrans Notification Webhook Engine', NOW() - INTERVAL '1 hour');
    END IF;

    IF user_admin IS NOT NULL THEN
        INSERT INTO audit_logs (id, actor_type, actor_id, action, entity_type, entity_id, ip_address, user_agent, created_at)
        VALUES 
            (gen_random_uuid(), 'admin', user_admin::text, 'CREATE_PROMO', 'promo', 'LOCKERIN50', '127.0.0.1', 'Lockerin Web Admin (Laravel/Inertia)', NOW() - INTERVAL '2 days');
    END IF;
END $$;
