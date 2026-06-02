CREATE TABLE uploads (
    id           BIGINT AUTO_INCREMENT PRIMARY KEY,
    fingerprint  VARCHAR(64) NOT NULL UNIQUE,
    video_name   VARCHAR(50) NOT NULL DEFAULT 'Untitled',
    video_id     VARCHAR(255) NOT NULL,
    upload_id    TEXT NOT NULL,        -- S3 multipart upload ID
    s3_key       TEXT NOT NULL,        -- where the file lives in S3
    part_size    INT NOT NULL,
    uploaded_parts JSON,
    status       VARCHAR(50),          -- uploading | completed | aborted
    created_at   TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at   TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

CREATE TABLE jobs (
    id           BIGINT AUTO_INCREMENT PRIMARY KEY,
    job_id       VARCHAR(255) NOT NULL UNIQUE,
    video_id     VARCHAR(255) NOT NULL,
    s3_key       TEXT NOT NULL,        -- source file key, needed by worker
    status       VARCHAR(50),          -- queued | downloading | transcoding | uploading | completed | failed
    stage        VARCHAR(50),
    created_at   TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at   TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);