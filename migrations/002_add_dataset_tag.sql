ALTER TABLE datasets ADD COLUMN tag VARCHAR(100) NULL;
CREATE INDEX idx_tag ON datasets (tag);
