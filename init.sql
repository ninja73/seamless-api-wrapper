CREATE TABLE IF NOT EXISTS currencies
(
    "id"   BIGSERIAL NOT NULL PRIMARY KEY,
    "code" VARCHAR   NOT NULL
);

INSERT INTO currencies(id, code)
VALUES (1, 'EUR');

CREATE TABLE IF NOT EXISTS balances
(
    "id"                          BIGSERIAL NOT NULL PRIMARY KEY,
    "player_name"                 VARCHAR   NOT NULL,
    "currency_id"                 BIGINT    NOT NULL,
    "amount"                      INTEGER   NOT NULL,
    "game_id"                     VARCHAR,
    "last_session_id"             VARCHAR,
    "last_session_alternative_id" VARCHAR,
    "free_round_left"             INTEGER,
    "created_at"                  TIMESTAMP WITHOUT TIME ZONE,
    "updated_at"                  TIMESTAMP WITHOUT TIME ZONE,
    CONSTRAINT
        fk_currency FOREIGN KEY (currency_id) REFERENCES currencies (id)
);

CREATE UNIQUE INDEX balances_player ON balances (player_name);

INSERT INTO balances(player_name, currency_id, amount, game_id, created_at, updated_at)
VALUES ('player1', 1, 10000, 'riot', NOW(), NOW());

CREATE TABLE IF NOT EXISTS transactions
(
    "id"                     BIGSERIAL NOT NULL PRIMARY KEY,
    "balance_id"             BIGINT    NOT NULL,
    "withdraw"               INTEGER   NOT NULL,
    "deposit"                INTEGER   NOT NULL,
    "transaction_ref"        VARCHAR   NOT NULL,
    "is_rollback"            BOOLEAN   NOT NULL DEFAULT FALSE,
    "game_id"                VARCHAR,
    "game_round_ref"         VARCHAR,
    "source"                 VARCHAR,
    "reason"                 VARCHAR,
    "session_id"             VARCHAR,
    "session_alternative_id" VARCHAR,
    "bonus_id"               VARCHAR,
    "charge_free_rounds"     INTEGER,
    "created_at"             TIMESTAMP WITHOUT TIME ZONE,
    "updated_at"             TIMESTAMP WITHOUT TIME ZONE,
    CONSTRAINT
        fk_balance FOREIGN KEY (balance_id) REFERENCES balances (id)
);

CREATE UNIQUE INDEX transactions_uniq_idx ON transactions (transaction_ref);
CREATE INDEX transactions_balance_idx ON transactions (balance_id);

CREATE TABLE IF NOT EXISTS spin_details
(
    "id"             BIGSERIAL NOT NULL PRIMARY KEY,
    "transaction_id" BIGINT    NOT NULL,
    "bet_type"       VARCHAR,
    "win_type"       VARCHAR,
    CONSTRAINT
        fk_transaction FOREIGN KEY (transaction_id) REFERENCES transactions (id)
);

CREATE INDEX spin_details_transaction_idx ON spin_details (transaction_id);