CREATE TABLE IF NOT EXISTS categories (
    id              SERIAL PRIMARY KEY,
    name            VARCHAR(100) NOT NULL UNIQUE,
    created_at      TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP 
);

CREATE TABLE IF NOT EXISTS products (
    id              SERIAL PRIMARY KEY,  
    name            VARCHAR(255) NOT NULL, 
    description     TEXT, 
    price           DECIMAL(10,2) NOT NULL,
    sku             VARCHAR(100) NOT NULL UNIQUE,
    image_url       VARCHAR(255) DEFAULT 'https://picsum.photos/seed/picsum/200/300',
    category_id     INTEGER REFERENCES categories(id),
    created_at      TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at      TIMESTAMP WITH TIME ZONE
);
