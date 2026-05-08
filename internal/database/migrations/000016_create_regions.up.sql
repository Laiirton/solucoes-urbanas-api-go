CREATE TABLE IF NOT EXISTS regions (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    neighborhoods JSONB NOT NULL DEFAULT '[]'::jsonb,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Add region_id to teams, remove old service_category
ALTER TABLE teams
ADD COLUMN IF NOT EXISTS region_id BIGINT REFERENCES regions(id) ON DELETE SET NULL;

ALTER TABLE teams DROP COLUMN IF EXISTS service_category;

-- Add region_id to service_requests
ALTER TABLE service_requests
ADD COLUMN IF NOT EXISTS region_id BIGINT REFERENCES regions(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_service_requests_region_id ON service_requests(region_id);
CREATE INDEX IF NOT EXISTS idx_teams_region_id ON teams(region_id);

-- Seed default regions
INSERT INTO regions (name, neighborhoods) VALUES
    ('Zona Norte', '["Tucuruvi", "Santana", "Mandaqui", "Vila Guilherme", "Vila Maria", "Jaçanã", "Tremembé", "Casa Verde", "Limão", "Freguesia do Ó", "Carandiru", "Imirim", "Lauzane Paulista"]'),
    ('Zona Sul', '["Jardins", "Moema", "Saúde", "Jabaquara", "Santo Amaro", "Campo Belo", "Campo Grande", "Vila Mariana", "Aclimação", "Paraíso", "Indianópolis", "Brooklin", "Vila Olímpia", "Itaim Bibi", "Cidade Jardim"]'),
    ('Centro', '["Sé", "Bela Vista", "Consolação", "República", "Santa Cecília", "Bom Retiro", "Liberdade", "Cambuci", "Higienópolis", "Vila Buarque", "Campos Elíseos", "Barra Funda"]'),
    ('Zona Leste', '["Tatuapé", "Mooca", "Penha", "Itaquera", "São Miguel", "Aricanduva", "Vila Matilde", "Artur Alvim", "Cidade Tiradentes", "Guaianases", "Ermelino Matarazzo", "Belenzinho", "Água Rasa", "Vila Prudente"]'),
    ('Zona Oeste', '["Pinheiros", "Butantã", "Perdizes", "Vila Leopoldina", "Lapa", "Alto de Pinheiros", "Jaguaré", "Vila Sônia", "Rio Pequeno", "Raposo Tavares", "Jardim Bonfiglioli", "Cidade Universitária"]')
ON CONFLICT DO NOTHING;
