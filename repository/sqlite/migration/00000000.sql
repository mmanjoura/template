CREATE TABLE user (
    Id         INTEGER PRIMARY KEY AUTOINCREMENT,
    [name]              [VARCHAR] (50),
    email             [VARCHAR] (50),
	api_key    TEXT NOT NULL UNIQUE,
    email_verified_at DATETIME,
    is_active         [INT],
    shop_Id           [INT],
    created_at                        TEXT NOT NULL,
    updated_at                        TEXT NOT NULL  
);


CREATE TABLE auth (
	Id            INTEGER PRIMARY KEY AUTOINCREMENT,
	user_Id       INTEGER NOT NULL REFERENCES user (Id) ON DELETE CASCADE,
	source        TEXT NOT NULL,
	source_Id     TEXT NOT NULL,
	access_token  TEXT NOT NULL,
	refresh_token TEXT NOT NULL,
	expiry        TEXT,
	created_at    TEXT NOT NULL,
	updated_at    TEXT NOT NULL,

	UNIQUE(user_Id, source),  -- one source per user
	UNIQUE(source, source_Id) -- one auth per source user
);

CREATE TABLE shop (
	Id          INTEGER PRIMARY KEY AUTOINCREMENT,
	user_Id     INTEGER NOT NULL REFERENCES user (Id) ON DELETE CASCADE,
	name        TEXT NOT NULL,
	invite_code TEXT UNIQUE NOT NULL,
	value       INTEGER NOT NULL DEFAULT 0,
	created_at  TEXT NOT NULL,
	updated_at  TEXT NOT NULL
);

CREATE INDEX shop_user_Id_Idx ON shop (user_Id);

CREATE TABLE shop_values (
	shop_Id      INTEGER NOT NULL REFERENCES shop (Id) ON DELETE CASCADE,
	"timestamp"  TEXT NOT NULL, -- per-minute precision
	value        INTEGER NOT NULL,

	PRIMARY KEY (shop_Id, "timestamp")
);

CREATE TABLE shop_membership (
	Id         INTEGER PRIMARY KEY AUTOINCREMENT,
	shop_Id    INTEGER NOT NULL REFERENCES shop (Id) ON DELETE CASCADE,
	user_Id    INTEGER NOT NULL REFERENCES user (Id) ON DELETE CASCADE,
	value      INTEGER NOT NULL,
	created_at TEXT NOT NULL,
	updated_at TEXT NOT NULL,

	UNIQUE(shop_Id, user_Id)
);

CREATE INDEX shop_membership_shop_Id_Idx ON shop_membership (shop_Id);
CREATE INDEX shop_membership_user_Id_Idx ON shop_membership (user_Id);

---------------------------- Start Inserts-----------------------------------------
INSERT INTO user([name], email, api_key, email_verified_at, is_active, shop_Id , created_at, updated_at ) VALUES ('Mustapha','mustapha.manjoura@gmail.com', '2d92763a84fd760444793a99432735378fce7e3d7af73fc213a81d3a08a9e715',  'null', 1, 1,  '2022-01-07T12:37:57Z', '2022-01-10T13:38:33Z');
INSERT INTO user([name], email, api_key, email_verified_at, is_active, shop_Id , created_at, updated_at ) VALUES ('mohamed','mohamed.manjoura@gmail.com', '521155290e1f40a40713f6f183e508486774175ca2f6ba3916d55fbf93fdbb5c',  'null', 1, 1, '2022-01-07T12:37:57Z', '2022-01-10T13:38:33Z');
INSERT INTO auth ( user_id, source, source_id, access_token, refresh_token, created_at, updated_at ) VALUES ( 1, 'github', 5604914, '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', '', '2022-01-07T12:37:57Z', '2022-01-10T13:38:33Z' );

