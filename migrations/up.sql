CREATE TABLE files (
    id SERIAL PRIMARY KEY,
    filename VARCHAR(150) UNIQUE NOT NULL,
    error VARCHAR(150) NOT NULL DEFAULT 'no error'
);

CREATE TABLE devices (
    n SERIAL NOT NULL,
    mqtt VARCHAR(100),
    invid VARCHAR(20) NOT NULL,
    unit_guid VARCHAR(50) NOT NULL,
    msg_id VARCHAR(100) NOT NULL,
    text TEXT NOT NULL,
    context VARCHAR(100),
    class VARCHAR(20) NOT NULL,
    level INTEGER NOT NULL,
    area VARCHAR(20) NOT NULL,
    addr VARCHAR(150) NOT NULL,
    block BOOLEAN,
    type VARCHAR(100),
    bit INTEGER,
    invert_bit BOOLEAN
);
