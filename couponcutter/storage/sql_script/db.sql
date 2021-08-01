DROP TABLE IF EXISTS users CASCADE;

DROP TABLE IF EXISTS stores CASCADE;

DROP TABLE IF EXISTS signup_users;

DROP TABLE IF EXISTS create_employees;

DROP Table IF EXISTS stores_employees CASCADE;

DROP TABLE IF EXISTS coupons CASCADE;

DROP TABLE IF EXISTS archive_coupons;

DROP TABLE IF EXISTS redeemed_coupons CASCADE;

DROP TABLE IF EXISTS coupon_categories CASCADE;

DROP TABLE IF EXISTS categories CASCADE;

DROP TABLE IF EXISTS employee_state CASCADE;

DROP TABLE IF EXISTS coupon_state CASCADE;

DROP TABLE IF EXISTS store_state CASCADE;

DROP TABLE IF EXISTS discount_type CASCADE;

CREATE TABLE employee_state(
    state_id PRIMARY KEY generated always AS IDENTITY,
    state_name text UNIQUE
);

CREATE TABLE coupon_state(
    state_id PRIMARY KEY generated always AS IDENTITY,
    state_name text UNIQUE
);

CREATE TABLE store_state(
    state_id PRIMARY KEY generated always AS IDENTITY,
    store_name text UNIQUE
);

CREATE TABLE discount_type(
    dis_id PRIMARY KEY generated always AS IDENTITY,
    dis_name text UNIQUE
);

INSERT INTO
    employee_state(state_name)
VALUES
    ('active'),
    ("removed"),
    ("suspended");

INSERT INTO
    coupon_state(state_name)
VALUES
    ('active'),
    ("deleted"),
    ("expired"),
    ("used");

INSERT INTO
    store_state(state_name)
VALUES
    ('active'),
    ("inactive");

INSERT INTO
    discount_type(dis_name)
VALUES
    ('amount_off'),
    ("percentage_off");

Create type store_state AS Enum('active', 'inactive');

CREATE TABLE users (
    user_id text PRIMARY KEY,
    email text NOT NULL UNIQUE,
    password_hash text NOT NULL,
    created_at timestamp NOT NULL
);

CREATE TABLE stores(
    store_id text PRIMARY KEY REFERENCES users(user_id),
    store_name text NOT NULL,
    store_state REFERENCES store_state(store_name) DEFAULT 'inactive',
    theme_color integer,
    tagline text NOT NULL,
    "address" text NULL
);

CREATE Table stores_followed(
    user_id text REFERENCES users(user_id) NOT NULL,
    store_id text REFERENCES stores(store_id) NOT NULL,
    UNIQUE(user_id, store_id)
);

Create TABLE signup_users(
    id integer PRIMARY KEY generated always AS IDENTITY,
    email text NOT NULL UNIQUE,
    password_hash text NOT NULL,
    token text NOT NULL UNIQUE,
    expired_at timestamp NOT NULL
);

Create Table create_employees(
    emp_id integer PRIMARY KEY generated always AS IDENTITY,
    store_id text REFERENCES stores(store_id),
    email text NOT NULL,
    token text NOT NULL UNIQUE,
    expired_at timestamp NOT NULL,
    revoked boolean NOT NULL DEFAULT false
);

CREATE TABLE stores_employees(
    emp_id text PRIMARY KEY,
    store_id text REFERENCES stores(store_id),
    created_at timestamp NOT NULL,
    user_id text REFERENCES users(user_id),
    emp_state REFERENCES employee_state(state_name)
);

Create TABLE coupons(
    coupon_id integer PRIMARY KEY generated always AS IDENTITY,
    store_id text REFERENCES stores(store_id) NOT NULL,
    "desc" text NOT NULL,
    "state" coupon_state NOT NULL DEFAULT 'active',
    discount_type text REFERENCES discount_type("dis_name") NOT NULL,
    amount_off NUMERIC,
    percentage_off NUMERIC,
    currency_code char(3),
    qr_code_url text NULL,
    expired_at timestamp NOT NULL,
    created_at timestamp NOT NULL,
    is_text_coupon boolean NOT NULL DEFAULT false,
    text_coupon_code text UNIQUE,
    text_coupon_weburl text,
    max_redemptions integer,
    unlimited_redemption boolean NOT NULL DEFAULT true,
    redemption_count integer NOT NULL DEFAULT 0,
    tsv tsvector,
    CONSTRAINT either_text_or_is_qr CHECK(
        is_text_coupon = false
        OR qr_code_url <> NULL
    ),
    CHECK(
        (
            unlimited_redemption <> true
            AND (redemption_count <= max_redemptions)
        )
        OR (unlimited_redemption)
    ),
    CHECK(
        amount_off <> NULL
        OR percentage_off <> NULL
    )
);

CREATE INDEX in_coupons_tsv ON coupons USING GIN(tsv);

CREATE
OR REPLACE FUNCTION set_full_text_search_on_coupons() RETURNS TRIGGER AS $$ BEGIN coupons.tsv = to_tsvector(coupons.desc ) RETURN NEW;

END;

$ $ language 'plpgsql';

CREATE TRIGGER update_new_full_text_search BEFORE
UPDATE
    ON coupons FOR EACH ROW EXECUTE PROCEDURE set_full_text_search_on_coupons();

Create TABLE saved_coupons(
    id integer PRIMARY KEY generated always AS IDENTITY,
    user_id text REFERENCES users(user_id) NOT NULL,
    coupon_id text REFERENCES coupons(coupon_id) NOT NULL,
    UNIQUE(user_id, coupon_id)
);

Create Table archive_coupons(
    coupon_id text PRIMARY KEY NOT NULL,
    store_id text REFERENCES stores(store_id) NOT NULL,
    "desc" text NOT NULL,
    coupon_state coupon_state NOT NULL,
    amount_off NUMERIC,
    percentage_off NUMERIC,
    currency_code char(3),
    qr_code_url text NULL,
    expired_at timestamp NOT NULL,
    created_at timestamp NOT NULL,
    is_text_coupon boolean NOT NULL DEFAULT false,
    text_coupon_code text UNIQUE,
    text_coupon_webUrl text,
    max_redemptions integer,
    unlimited_redemption boolean NOT NULL DEFAULT true,
    redemption_count integer NOT NULL DEFAULT 0,
    CHECK(coupon_state = 'deleted')
);

CREATE TABLE redeemed_coupons(
    id integer PRIMARY KEY generated always AS IDENTITY,
    coupon_id text REFERENCES coupons(coupon_id) NOT NULL,
    redeemed_by text REFERENCES stores_employees(employee_id),
    redeemed_when timestamp NOT NULL
);

CREATE TABLE categories (
    cat_id integer PRIMARY KEY generated always AS IDENTITY,
    cat_name text NOT NULL
);

CREATE TABLE coupon_categories(
    coupon_id text REFERENCES coupons(coupon_id),
    cat_name integer REFERENCES categories(cat_name),
    PRIMARY KEY(coupon_id, cat_id)
);

CREATE MATERIALIZED VIEW popular_coupons As
SELECT
    coupon_id,
    store_id,
    store_name,
    tagline,
    "address",
    theme_color,
    "desc",
    amount_off,
    percentage_off,
    currency_code,
    qr_code_url,
    expired_at,
    is_text_coupon,
    text_coupon_code,
    text_coupon_webUrl,
    redeemed_count
FROM
    (
        SELECT
            coupon_id,
            amount_off,
            percentage_off,
            currency_code,
            qr_code_url,
            "desc",
            expired_at,
            is_text_coupon,
            text_coupon_code,
            text_coupon_webUrl,
            store_id,
            count(coupon_id) AS redeemed_count
        FROM
            coupons
            INNER JOIN redeemed_coupons using(coupon_id)
        WHERE
            coupon_state = 'active'
        GROUP BY
            coupon_id
        ORDER BY
            redeemed_count
    ) AS redeemed
    INNER JOIN stores using(store_id);