CREATE TABLE IF NOT EXISTS labels (
                                      exchange   TEXT NOT NULL,
                                      symbol     TEXT NOT NULL,
                                      timeframe  TEXT NOT NULL,
                                      timestamp  INTEGER NOT NULL,

                                      fwd_ret    REAL NOT NULL,
                                      label      TEXT NOT NULL,
                                      threshold_b REAL NOT NULL,

                                      PRIMARY KEY (exchange, symbol, timeframe, timestamp),
    CHECK (label IN ('UP','DOWN','NO_TRADE'))
    );