CREATE TABLE users (
    id                          INTEGER PRIMARY KEY,
    username                    TEXT NOT NULL UNIQUE,
    admin                       BOOLEAN NOT NULL,
    author                      BOOLEAN NOT NULL,
    salt                        BLOB NOT NULL,
    scheme                      TEXT NOT NULL,
    password_hash               BLOB NOT NULL,
    last_signed_in_at           DATETIME NOT NULL,
    created_at                  DATETIME NOT NULL,
    modified_at                 DATETIME NOT NULL
);

CREATE TABLE accounts (
    id                          INTEGER PRIMARY KEY,
    user_id                     INTEGER NOT NULL,
    last_signed_in_at           DATETIME NOT NULL,
    created_at                  DATETIME NOT NULL,
    modified_at                 DATETIME NOT NULL,

    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE ON UPDATE CASCADE
);

CREATE TABLE sessions (
    id                          INTEGER PRIMARY KEY,
    user_id                     INTEGER NOT NULL,
    signed_in_from              TEXT NOT NULL,
    signed_in_at                DATETIME NOT NULL,
    expires_at                  DATETIME NOT NULL,

    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE ON UPDATE RESTRICT
);

CREATE TABLE areas (
    id                          INTEGER PRIMARY KEY,
    name                        TEXT NOT NULL,
    created_at                  DATETIME NOT NULL,
    modified_at                 DATETIME NOT NULL
);

CREATE TABLE helps (
    id                          INTEGER PRIMARY KEY,
    area_id                     INTEGER NOT NULL,
    level                       INTEGER NOT NULL,
    keywords                    TEXT NOT NULL,
    help_text                   TEXT NOT NULL,

    FOREIGN KEY (area_id) REFERENCES areas (id) ON DELETE CASCADE ON UPDATE CASCADE,
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
    pronouns                    TEXT NOT NULL,

    FOREIGN KEY (area_id) REFERENCES areas (id) ON DELETE CASCADE ON UPDATE CASCADE,
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
    extras                      TEXT NOT NULL,
    applies                     TEXT NOT NULL,

    FOREIGN KEY (area_id) REFERENCES areas (id) ON DELETE CASCADE ON UPDATE CASCADE
);

CREATE TABLE rooms (
    id                          INTEGER PRIMARY KEY,
    area_id                     INTEGER NOT NULL,
    name                        TEXT NOT NULL,
    description                 TEXT NOT NULL,
    flags                       INTEGER NOT NULL,
    terrain                     INTEGER NOT NULL,
    extras                      TEXT NOT NULL,

    FOREIGN KEY (area_id) REFERENCES areas (id) ON DELETE CASCADE ON UPDATE CASCADE
);

CREATE TABLE doors (
	id							INTEGER PRIMARY KEY,
	room_id						INTEGER NOT NULL,
	direction					INTEGER NOT NULL,
	description					TEXT NOT NULL,
	keywords 					TEXT NOT NULL,
	lock						INTEGER,
	key							INTEGER,
	to_room						INTEGER,

	FOREIGN KEY (room_id) REFERENCES rooms (id) ON DELETE CASCADE ON UPDATE CASCADE,
	FOREIGN KEY (to_room) REFERENCES rooms (id) ON DELETE CASCADE ON UPDATE CASCADE
);

CREATE TABLE resets (
    id                          INTEGER PRIMARY KEY,
    reset_type                  TEXT NOT NULL,
    area_id                     INTEGER NOT NULL,
    sequence                    INTEGER NOT NULL,
    room_id                     INTEGER,
    mobile_id                   INTEGER,
    object_id                   INTEGER,
    container_id                INTEGER,
    wear_location               INTEGER,
    max_instances               INTEGER,
    door_direction              INTEGER,
    door_state                  INTEGER,
    last_door                   INTEGER,
    comment                     TEXT,

    FOREIGN KEY (area_id) REFERENCES areas (id) ON DELETE CASCADE ON UPDATE CASCADE,
    FOREIGN KEY (room_id) REFERENCES rooms (id) ON DELETE CASCADE ON UPDATE CASCADE,
    FOREIGN KEY (mobile_id) REFERENCES mobiles (id) ON DELETE CASCADE ON UPDATE CASCADE,
    FOREIGN KEY (object_id) REFERENCES objects (id) ON DELETE CASCADE ON UPDATE CASCADE,
    FOREIGN KEY (container_id) REFERENCES objects (id) ON DELETE CASCADE ON UPDATE CASCADE
);
