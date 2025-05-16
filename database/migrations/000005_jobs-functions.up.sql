-- Başvuru sayısını güncelleyen fonksiyon (INSERT)
CREATE OR REPLACE FUNCTION update_applicants_count()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE job_posting_details
    SET applicants = (
        SELECT COUNT(*)
        FROM job_applications
        WHERE job_id = NEW.job_id
    )
    WHERE id = NEW.job_id;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- Başvuru sayısını güncelleyen fonksiyon (DELETE)
CREATE OR REPLACE FUNCTION update_applicants_count_on_delete()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE job_posting_details
    SET applicants = (
        SELECT COUNT(*)
        FROM job_applications
        WHERE job_id = OLD.job_id
    )
    WHERE id = OLD.job_id;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- Başvuru sayısı için trigger'lar
CREATE TRIGGER trg_applicants_after_insert
AFTER INSERT ON job_applications
FOR EACH ROW
EXECUTE FUNCTION update_applicants_count();

CREATE TRIGGER trg_applicants_after_delete
AFTER DELETE ON job_applications
FOR EACH ROW
EXECUTE FUNCTION update_applicants_count_on_delete();

-- Update_At için trigger'lar.
CREATE TRIGGER update_job_postings_updated_at
BEFORE UPDATE ON job_postings
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_job_posting_details_updated_at
BEFORE UPDATE ON job_posting_details
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_job_categories_updated_at
BEFORE UPDATE ON job_categories
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_job_applications_updated_at
BEFORE UPDATE ON job_applications
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();
