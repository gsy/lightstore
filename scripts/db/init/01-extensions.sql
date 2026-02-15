-- Enable useful PostgreSQL extensions

-- UUID generation
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Case-insensitive text
CREATE EXTENSION IF NOT EXISTS "citext";

-- Trigram matching for fuzzy search
CREATE EXTENSION IF NOT EXISTS "pg_trgm";
