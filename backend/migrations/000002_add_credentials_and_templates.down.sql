-- Drop workflow_templates table
DROP INDEX IF EXISTS idx_workflow_templates_created_at;
DROP INDEX IF EXISTS idx_workflow_templates_featured;
DROP INDEX IF EXISTS idx_workflow_templates_category;
DROP TABLE IF EXISTS workflow_templates;

-- Drop credentials table
DROP INDEX IF EXISTS idx_credentials_created_at;
DROP INDEX IF EXISTS idx_credentials_type;
DROP TABLE IF EXISTS credentials;
