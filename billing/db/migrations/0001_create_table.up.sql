CREATE TYPE bill_status AS ENUM ('open', 'closed');
CREATE TYPE currency_type AS ENUM ('GEL', 'USD');

CREATE TABLE bill (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  status bill_status NOT NULL DEFAULT 'open',
  created_at TIMESTAMP DEFAULT now(),
  close_date TIMESTAMP NOT NULL,
  closed_at TIMESTAMP NULL
);

CREATE TABLE bill_item (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  bill_id UUID NOT NULL,
  amount INT NOT NULL,
  created_at TIMESTAMP DEFAULT now(),
  currency currency_type NOT NULL,

  FOREIGN KEY (bill_id) REFERENCES bill(id) ON DELETE CASCADE
);

CREATE VIEW bill_summary AS
SELECT
  bill_id,
  currency,
  SUM(amount) as total_amount
FROM
  bill_item
GROUP BY
  bill_id, currency;