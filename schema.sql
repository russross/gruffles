CREATE TABLE users (
    id                          INTEGER PRIMARY KEY,
    username                    TEXT NOT NULL,
    salt                        TEXT NOT NULL,
    scheme                      TEXT NOT NULL,
    password_hash               TEXT NOT NULL,
    last_signed_in_at           TEXT NOT NULL,
    created_at                  TEXT NOT NULL,
    modified_at                 TEXT NOT NULL
);

CREATE TABLE accounts (
    id                          INTEGER PRIMARY KEY,
    user_id                     INTEGER NOT NULL,
    last_signed_in_at           TEXT NOT NULL,
    created_at                  TEXT NOT NULL,
    modified_at                 TEXT NOT NULL,

    FOREIGN KEY (user_id) REFERENCES users (id)
);

CREATE TABLE sessions (
    id                          INTEGER PRIMARY KEY,
    user_id                     INTEGER NOT NULL,
    key                         TEXT NOT NULL UNIQUE,
    signed_in_at                TEXT NOT NULL,
    expires_at                  TEXT NOT NULL,

    FOREIGN KEY (user_id) REFERENCES users (id)
);

CREATE TABLE areas (
    id                          INTEGER PRIMARY KEY,
    name                        TEXT NOT NULL,
    created_at                  TEXT NOT NULL,
    modified_at                 TEXT NOT NULL
);

CREATE TABLE helps (
    id                          INTEGER PRIMARY KEY,
    area_id                     INTEGER NOT NULL,
    level                       INTEGER NOT NULL,
    keywords                    TEXT NOT NULL,
    help_text                   TEXT NOT NULL,

    FOREIGN KEY (area_id) REFERENCES areas (id),
	CHECK (level >= 0 AND level <= 100)
);

CREATE TABLE mobiles (
    id                          INTEGER PRIMARY KEY,
    area_id                     INTEGER NOT NULL,
    keywords                    TEXT NOT NULL,
    short_description           TEXT NOT NULL,
    long_description            TEXT NOT NULL,
    description                 TEXT NOT NULL,
    action_flags                INTEGER NOT NULL,
    affected_flags              INTEGER NOT NULL,
    alignment                   INTEGER NOT NULL,
    level                       INTEGER NOT NULL,
    hit_roll                    TEXT NOT NULL,
    damage_roll                 TEXT NOT NULL,
    dodge_roll                  TEXT NOT NULL,
    absorb_roll                 TEXT NOT NULL,
    fire_roll                   TEXT NOT NULL,
    ice_roll                    TEXT NOT NULL,
    poison_roll                 TEXT NOT NULL,
    lightning_roll              TEXT NOT NULL,
    gold                        INTEGER NOT NULL,
    experience                  INTEGER NOT NULL,
    pronouns                    text NOT NULL,

    FOREIGN KEY (area_id) REFERENCES areas (id),
	CHECK (alignment >= -1000 AND alignment <= 1000),
	CHECK (level >= 0 AND level <= 100),
	CHECK (pronouns IN ("he", "she", "it", "they"))
);

CREATE TABLE objects (
    id                          INTEGER PRIMARY KEY,
    area_id                     INTEGER NOT NULL,
    keywords                    TEXT NOT NULL,
    short_description           TEXT NOT NULL,
    long_description            TEXT NOT NULL,
    item_type                   INTEGER NOT NULL,
    extra_flags                 INTEGER NOT NULL,
    wear_flags                  INTEGER NOT NULL,
    value_0                     INTEGER NOT NULL,
    value_1                     INTEGER NOT NULL,
    value_2                     INTEGER NOT NULL,
    value_3                     INTEGER NOT NULL,
    weight                      INTEGER NOT NULL,
    cost                        INTEGER NOT NULL,

    FOREIGN KEY (area_id) REFERENCES areas (id)
);

CREATE TABLE object_extra_descriptions (
    id                          INTEGER PRIMARY KEY,
    object_id                   INTEGER NOT NULL,
    keywords                    TEXT NOT NULL,
    description                 TEXT NOT NULL,

    FOREIGN KEY (object_id) REFERENCES objects (id)
);

CREATE TABLE object_applies (
    id                          INTEGER PRIMARY KEY,
    object_id                   INTEGER NOT NULL,
    apply_type                  INTEGER NOT NULL,
    value                       INTEGER NOT NULL,

    FOREIGN KEY (object_id) REFERENCES objects (id)
);

CREATE TABLE rooms (
    id                          INTEGER PRIMARY KEY,
    area_id                     INTEGER NOT NULL,
    name                        TEXT NOT NULL,
    description                 TEXT NOT NULL,
    flags                       INTEGER NOT NULL,
    terrain                     INTEGER NOT NULL,

    FOREIGN KEY (area_id) REFERENCES areas (id)
);

CREATE TABLE doors (
    id                          INTEGER PRIMARY KEY,
    room_id                     INTEGER NOT NULL,
    direction                   INTEGER NOT NULL,
    description                 TEXT NOT NULL,
    keywords                    TEXT NOT NULL,
    lock                        INTEGER NOT NULL,
    key                         INTEGER NOT NULL,
    to_room                     INTEGER NOT NULL,

    FOREIGN KEY (room_id) REFERENCES rooms (id),
    FOREIGN KEY (to_room) REFERENCES rooms (id)
);

CREATE TABLE room_extra_descriptions (
    id                          INTEGER PRIMARY KEY,
    room_id                     INTEGER NOT NULL,
    keywords                    TEXT NOT NULL,
    description                 TEXT NOT NULL,

    FOREIGN KEY (room_id) REFERENCES rooms (id)
);

CREATE TABLE resets (
    id                          INTEGER PRIMARY KEY,
    reset_type                  INTEGER NOT NULL,
    area_id                     INTEGER NOT NULL,
    room_id                     INTEGER,
    mobile_id                   INTEGER,
    object_id                   INTEGER,
    container_id                INTEGER,
    max_instances               INTEGER,
    door_direction              INTEGER,
    door_state                  INTEGER,

    FOREIGN KEY (area_id) REFERENCES areas (id),
    FOREIGN KEY (room_id) REFERENCES rooms (id),
    FOREIGN KEY (mobile_id) REFERENCES mobiles (id),
    FOREIGN KEY (object_id) REFERENCES objects (id),
    FOREIGN KEY (container_id) REFERENCES objects (id)
);
