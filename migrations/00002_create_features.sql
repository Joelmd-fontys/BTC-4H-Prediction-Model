CREATE TABLE IF NOT EXISTS features (
                                        exchange   TEXT NOT NULL,
                                        symbol     TEXT NOT NULL,
                                        timeframe  TEXT NOT NULL,
                                        timestamp  INTEGER NOT NULL,

                                        ret_1      REAL,
                                        vol_20     REAL,
                                        mom_6      REAL,
                                        ema_10     REAL,
                                        ema_30     REAL,
                                        ema_spread REAL,
                                        range_hl   REAL,
                                        range_co   REAL,
                                        vol_chg    REAL,

                                        PRIMARY KEY (exchange, symbol, timeframe, timestamp)
    );