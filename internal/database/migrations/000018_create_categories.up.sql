CREATE TABLE IF NOT EXISTS categories (
    id          BIGSERIAL PRIMARY KEY,
    name        VARCHAR   NOT NULL UNIQUE,
    icon        VARCHAR,
    is_active   BOOLEAN   NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ DEFAULT NOW(),
    updated_at  TIMESTAMPTZ DEFAULT NOW()
);

-- Seed with existing categories from services table
INSERT INTO categories (name, icon, is_active)
SELECT DISTINCT 
    category as name,
    CASE category
        WHEN 'Limpeza Urbana' THEN 'brush-cleaning'
        WHEN 'Saúde' THEN 'hospital'
        WHEN 'Educação' THEN 'school'
        WHEN 'Iluminação Pública' THEN 'lightbulb'
        WHEN 'Transporte Urbano' THEN 'bus'
        WHEN 'Segurança Pública' THEN 'shield'
        WHEN 'Esporte e Lazer' THEN 'bike'
        WHEN 'Cultura' THEN 'theater'
        WHEN 'Tributação' THEN 'hand-coins'
        WHEN 'Assistência Social' THEN 'hand-helping'
        WHEN 'Vias Urbanas' THEN 'arrow-left-right'
        WHEN 'Arborização e Meio Ambiente' THEN 'tree'
        WHEN 'Agricultura' THEN 'sprout'
        WHEN 'Vigilância Sanitária' THEN 'shield'
        WHEN 'Animais' THEN 'paw'
        ELSE 'help-circle'
    END as icon,
    TRUE as is_active
FROM services
ON CONFLICT (name) DO NOTHING;
