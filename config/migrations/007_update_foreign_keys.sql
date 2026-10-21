ALTER TABLE user_sessions
    DROP CONSTRAINT IF EXISTS user_sessions_user_id_fkey,
    ADD CONSTRAINT user_sessions_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id);

ALTER TABLE fields
    DROP CONSTRAINT IF EXISTS fields_page_id_fkey,
    ADD CONSTRAINT fields_page_id_fkey FOREIGN KEY (page_id) REFERENCES pages(id);
