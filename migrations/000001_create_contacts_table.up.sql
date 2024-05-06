CREATE TABLE IF NOT EXISTS contacts (
    id bigserial PRIMARY KEY,  
    created timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    first text NOT NULL,
    last text NOT NULL,
    phone text NOT NULL,
    email text NOT NULL,
    version integer NOT NULL DEFAULT 1
);