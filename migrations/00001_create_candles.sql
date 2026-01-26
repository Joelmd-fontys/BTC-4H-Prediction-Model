create table candles(
                        exchange TEXT NOT NULL,
                        symbol   TEXT NOT NULL,
                        timeframe TEXT NOT NULL,
                        timestamp       INTEGER NOT NULL,
                        open  REAL NOT NULL,
                        high  REAL NOT NULL,
                        low   REAL NOT NULL,
                        close REAL NOT NULL,
                        volume REAL NOT NULL,
                        close_time INTEGER,
                        is_final INTEGER NOT NULL DEFAULT 1,
                        PRIMARY KEY (exchange, symbol, timeframe, timestamp),
                        CHECK (volume >= 0),
                        CHECK (low <= high),
                        CHECK (low <= open AND low <= close),
                        CHECK (high >= open AND high >= close),
                        CHECK (is_final IN (0,1)),
                        CHECK (close_time IS NULL OR close_time > timestamp),
                        CHECK (timestamp > 0)
)

