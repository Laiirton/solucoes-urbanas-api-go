-- Add form_schema column to services table
-- This allows admins to define the dynamic fields for each service type

ALTER TABLE services ADD COLUMN IF NOT EXISTS form_schema JSONB NOT NULL DEFAULT '[]'::jsonb;
