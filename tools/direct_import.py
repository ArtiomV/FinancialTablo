#!/usr/bin/env python3
"""
Direct import of xlsx transactions into ezbookkeeping SQLite database.
Skips the web import pipeline entirely.
"""

import openpyxl
import sqlite3
import time
from datetime import datetime, timezone, timedelta

# ── Config ──────────────────────────────────────────────────────────────────
XLSX_PATH = '/tmp/test_import.xlsx'
DB_PATH = '/root/ezbookkeeping-data/ezbookkeeping.db'
UID = 3803138511473737728
UUID_SERVER_ID = 0
TIMEZONE_OFFSET_MINUTES = 180  # Moscow UTC+3 (ezbookkeeping uses positive for east of UTC)

UUID_TYPE_TRANSACTION = 3
UUID_TYPE_TAG_INDEX = 6

TX_TYPE_INCOME = 2
TX_TYPE_EXPENSE = 3
TX_TYPE_TRANSFER_OUT = 4
TX_TYPE_TRANSFER_IN = 5

TRANSFER_CATEGORIES = {'Конвертация валют', 'Перевод между счетами'}
MOSCOW = timezone(timedelta(hours=3))


class UuidGenerator:
    SEQ_ID_MASK = (1 << 19) - 1

    def __init__(self, server_id=0):
        self.server_id = server_id
        self.seq_counters = {}

    def generate(self, uuid_type, unix_time=None):
        if unix_time is None:
            unix_time = int(time.time())
        key = (uuid_type, unix_time)
        if key in self.seq_counters:
            self.seq_counters[key] += 1
        else:
            self.seq_counters[key] = 0
        seq = self.seq_counters[key]
        uid = (unix_time & ((1 << 32) - 1)) << (4 + 8 + 19)
        uid |= (uuid_type & 0xF) << (8 + 19)
        uid |= (self.server_id & 0xFF) << 19
        uid |= (seq & self.SEQ_ID_MASK)
        return uid


def parse_date(date_val):
    """Parse date to unix timestamp in MILLISECONDS at 00:00 Moscow time"""
    if isinstance(date_val, datetime):
        dt = datetime(date_val.year, date_val.month, date_val.day, 0, 0, 0, tzinfo=MOSCOW)
        return int(dt.timestamp()) * 1000

    s = str(date_val).strip()
    # Handle "YYYY-MM-DD HH:MM:SS" format from openpyxl
    if ' ' in s:
        s = s.split(' ')[0]
    if '-' in s and len(s.split('-')[0]) == 4:
        parts = s.split('-')
        dt = datetime(int(parts[0]), int(parts[1]), int(parts[2]), 0, 0, 0, tzinfo=MOSCOW)
        return int(dt.timestamp()) * 1000

    parts = s.split('.')
    if len(parts) == 3:
        day, month, year = int(parts[0]), int(parts[1]), int(parts[2])
        dt = datetime(year, month, day, 0, 0, 0, tzinfo=MOSCOW)
        return int(dt.timestamp()) * 1000

    raise ValueError(f"Bad date: {date_val}")


def parse_amount(amt_val):
    """Parse amount. Returns (abs_cents, is_negative)."""
    s = str(amt_val).strip()
    if not s:
        return 0, False

    is_negative = False
    if s.startswith('(') and s.endswith(')'):
        s = s[1:-1].strip()
    if s.startswith('-'):
        is_negative = True
        s = s[1:].strip()

    s = s.replace(' ', '').replace('\xa0', '').replace(',', '.')
    value = float(s)
    cents = int(round(value * 100))
    return cents, is_negative


def main():
    print("Loading xlsx...")
    wb = openpyxl.load_workbook(XLSX_PATH)
    ws = wb.active
    all_rows = list(ws.iter_rows(min_row=2, values_only=True))
    print(f"Total rows: {len(all_rows)}")

    db = sqlite3.connect(DB_PATH)
    cur = db.cursor()

    # Load reference data
    cur.execute('SELECT account_id, name, currency FROM account WHERE uid=? AND deleted=0', (UID,))
    accounts = {r[1]: {'id': r[0], 'currency': r[2]} for r in cur.fetchall()}

    cur.execute('SELECT category_id, name, type FROM transaction_category WHERE uid=? AND deleted=0', (UID,))
    categories = {}
    for r in cur.fetchall():
        categories[(r[1], r[2])] = r[0]

    cur.execute('SELECT tag_id, name, tag_group_id FROM transaction_tag WHERE uid=? AND deleted=0', (UID,))
    tags = {r[1]: {'id': r[0], 'group_id': r[2]} for r in cur.fetchall()}

    cur.execute('SELECT counterparty_id, name FROM counterparty WHERE uid=? AND deleted=0', (UID,))
    counterparties = {}
    for r in cur.fetchall():
        counterparties[r[1]] = r[0]
        # Also store normalized (no newlines)
        normalized = r[1].replace('\n', ' ').strip()
        if normalized not in counterparties:
            counterparties[normalized] = r[0]

    # ── Parse all rows ──
    parsed_rows = []
    skip_count = 0
    for i, row in enumerate(all_rows):
        date_val, amt_val = row[0], row[1]
        account_name = str(row[2]).strip() if row[2] else None
        category_name = str(row[6]).strip() if row[6] else None
        counterparty_name = str(row[4]).strip().replace('\n', ' ') if row[4] else None
        description = str(row[8]).strip() if row[8] else ''
        direction = str(row[9]).strip() if row[9] else None
        subdirection = str(row[10]).strip() if row[10] else None

        if not date_val or not amt_val or not account_name or not category_name:
            skip_count += 1
            continue

        try:
            tx_time = parse_date(date_val)
        except Exception as e:
            skip_count += 1
            continue

        try:
            abs_cents, is_negative = parse_amount(amt_val)
        except:
            skip_count += 1
            continue

        if category_name in TRANSFER_CATEGORIES:
            tx_type = TX_TYPE_TRANSFER_OUT
        elif is_negative:
            tx_type = TX_TYPE_EXPENSE
        else:
            tx_type = TX_TYPE_INCOME

        # Category lookup
        if tx_type == TX_TYPE_INCOME:
            cat_key = (category_name, 1)
        elif tx_type == TX_TYPE_EXPENSE:
            cat_key = (category_name, 2)
        else:
            cat_key = (category_name, 3)

        category_id = categories.get(cat_key)
        if not category_id:
            for t in [1, 2, 3]:
                if (category_name, t) in categories:
                    category_id = categories[(category_name, t)]
                    break
        if not category_id:
            skip_count += 1
            continue

        account_info = accounts.get(account_name)
        if not account_info:
            skip_count += 1
            continue

        counterparty_id = 0
        if counterparty_name:
            counterparty_id = counterparties.get(counterparty_name, 0)

        row_tags = []
        if direction and direction in tags:
            t = tags[direction]
            row_tags.append(t['id'])
        if subdirection and subdirection in tags:
            t = tags[subdirection]
            row_tags.append(t['id'])

        parsed_rows.append({
            'tx_time': tx_time,
            'tx_type': tx_type,
            'category_id': category_id,
            'category_name': category_name,
            'account_id': account_info['id'],
            'amount_cents': abs_cents,
            'is_negative': is_negative,
            'counterparty_id': counterparty_id,
            'comment': description,
            'tags': row_tags,
        })

    print(f"Parsed: {len(parsed_rows)}, skipped: {skip_count}")

    # ── Assign unique timestamps ──
    # UNIQUE constraint on (uid, transaction_time), so each tx needs unique time.
    # Group by base time, then add 1 sec offset for each within same time.
    time_counter = {}
    for row in parsed_rows:
        base = row['tx_time']
        if base in time_counter:
            time_counter[base] += 1
        else:
            time_counter[base] = 0
        row['unique_time'] = base + time_counter[base]

    # Check for collisions in unique_time
    seen_times = set()
    for row in parsed_rows:
        while row['unique_time'] in seen_times:
            row['unique_time'] += 1
        seen_times.add(row['unique_time'])

    # ── Match transfer pairs ──
    transfer_rows = [r for r in parsed_rows if r['tx_type'] == TX_TYPE_TRANSFER_OUT]
    non_transfer_rows = [r for r in parsed_rows if r['tx_type'] != TX_TYPE_TRANSFER_OUT]

    neg_transfers = [r for r in transfer_rows if r['is_negative']]
    pos_transfers = [r for r in transfer_rows if not r['is_negative']]

    transfer_pairs = []
    used_pos = set()

    for neg in neg_transfers:
        matched = None
        for j, pos in enumerate(pos_transfers):
            if j in used_pos:
                continue
            # Match: same base date (tx_time before offset), same category
            if neg['tx_time'] == pos['tx_time'] and neg['category_name'] == pos['category_name']:
                matched = pos
                used_pos.add(j)
                break
        transfer_pairs.append((neg, matched))

    for j, pos in enumerate(pos_transfers):
        if j not in used_pos:
            transfer_pairs.append((None, pos))

    matched = sum(1 for s, d in transfer_pairs if s and d)
    print(f"Transfers: {len(transfer_rows)} rows, {matched} matched pairs, {len(transfer_pairs) - matched} unmatched")

    # ── Insert into DB ──
    gen = UuidGenerator(UUID_SERVER_ID)
    base_gen_time = int(time.time())
    now_unix = base_gen_time
    tx_count = 0
    tag_count = 0

    def insert_tx(tx_id, tx_type, category_id, account_id, unique_time, amount, comment,
                  counterparty_id, related_id=0, related_account_id=0, related_account_amount=0):
        cur.execute('''INSERT INTO "transaction" (
            transaction_id, uid, deleted, type, category_id, account_id,
            transaction_time, timezone_utc_offset, amount,
            related_id, related_account_id, related_account_amount,
            hide_amount, comment, counterparty_id,
            geo_longitude, geo_latitude, created_ip, scheduled_created,
            planned, source_template_id,
            created_unix_time, updated_unix_time, deleted_unix_time, cfo_id
        ) VALUES (?,?,0,?,?,?,?,?,?,?,?,?,0,?,?,0,0,NULL,NULL,0,0,?,?,0,0)''',
        (tx_id, UID, tx_type, category_id, account_id,
         unique_time, TIMEZONE_OFFSET_MINUTES, amount,
         related_id, related_account_id, related_account_amount,
         comment, counterparty_id, now_unix, now_unix))

    def insert_tags(tx_id, unique_time, tag_ids):
        nonlocal tag_count
        for tag_id in tag_ids:
            ti_id = gen.generate(UUID_TYPE_TAG_INDEX, base_gen_time)
            cur.execute('''INSERT INTO transaction_tag_index (
                tag_index_id, uid, deleted, transaction_time, tag_id, transaction_id,
                created_unix_time, updated_unix_time, deleted_unix_time
            ) VALUES (?,?,0,?,?,?,?,?,0)''',
            (ti_id, UID, unique_time, tag_id, tx_id, now_unix, now_unix))
            tag_count += 1

    # 1. Non-transfer transactions
    for row in non_transfer_rows:
        tx_id = gen.generate(UUID_TYPE_TRANSACTION, base_gen_time)
        insert_tx(tx_id, row['tx_type'], row['category_id'], row['account_id'],
                  row['unique_time'], row['amount_cents'], row['comment'], row['counterparty_id'])
        insert_tags(tx_id, row['unique_time'], row['tags'])
        tx_count += 1

    # 2. Transfer pairs
    for source, dest in transfer_pairs:
        if source and dest:
            tx_id_src = gen.generate(UUID_TYPE_TRANSACTION, base_gen_time)
            tx_id_dst = gen.generate(UUID_TYPE_TRANSACTION, base_gen_time)

            comment = source['comment'] or dest['comment']
            cp_id = source['counterparty_id'] or dest['counterparty_id']

            insert_tx(tx_id_src, TX_TYPE_TRANSFER_OUT, source['category_id'], source['account_id'],
                      source['unique_time'], source['amount_cents'], comment, cp_id,
                      related_id=tx_id_dst, related_account_id=dest['account_id'],
                      related_account_amount=dest['amount_cents'])

            insert_tx(tx_id_dst, TX_TYPE_TRANSFER_IN, dest['category_id'], dest['account_id'],
                      dest['unique_time'], dest['amount_cents'], comment, cp_id,
                      related_id=tx_id_src, related_account_id=source['account_id'],
                      related_account_amount=source['amount_cents'])

            tx_count += 2

            all_tags = list(set(source['tags'] + dest['tags']))
            insert_tags(tx_id_src, source['unique_time'], all_tags)
            insert_tags(tx_id_dst, dest['unique_time'], all_tags)

        elif source:
            tx_id = gen.generate(UUID_TYPE_TRANSACTION, base_gen_time)
            insert_tx(tx_id, TX_TYPE_TRANSFER_OUT, source['category_id'], source['account_id'],
                      source['unique_time'], source['amount_cents'], source['comment'],
                      source['counterparty_id'])
            insert_tags(tx_id, source['unique_time'], source['tags'])
            tx_count += 1

        elif dest:
            tx_id = gen.generate(UUID_TYPE_TRANSACTION, base_gen_time)
            insert_tx(tx_id, TX_TYPE_TRANSFER_OUT, dest['category_id'], dest['account_id'],
                      dest['unique_time'], dest['amount_cents'], dest['comment'],
                      dest['counterparty_id'])
            insert_tags(tx_id, dest['unique_time'], dest['tags'])
            tx_count += 1

    db.commit()

    # ── Verify ──
    cur.execute('SELECT COUNT(*) FROM "transaction" WHERE uid=? AND deleted=0', (UID,))
    total = cur.fetchone()[0]

    cur.execute('SELECT type, COUNT(*) FROM "transaction" WHERE uid=? AND deleted=0 GROUP BY type', (UID,))
    by_type = dict(cur.fetchall())

    cur.execute('SELECT COUNT(*) FROM transaction_tag_index WHERE uid=? AND deleted=0', (UID,))
    total_tags = cur.fetchone()[0]

    # Verify the -32000 google-аккаунтов row
    target = int(datetime(2021, 7, 6, 0, 0, 0, tzinfo=MOSCOW).timestamp()) * 1000
    cur.execute('''SELECT transaction_id, amount, comment, transaction_time
                   FROM "transaction" WHERE uid=? AND deleted=0
                   AND transaction_time >= ? AND transaction_time < ?
                   AND comment LIKE "%google%"''',
                (UID, target, target + 86400))
    google = cur.fetchall()

    print(f"\n{'='*60}")
    print(f"IMPORT COMPLETE!")
    print(f"{'='*60}")
    print(f"Transactions inserted: {tx_count}")
    print(f"Tag indexes inserted: {tag_count}")
    print(f"Total in DB: {total}")
    print(f"  Income:   {by_type.get(2, 0)}")
    print(f"  Expense:  {by_type.get(3, 0)}")
    print(f"  Transfer: {by_type.get(4, 0)}")
    print(f"Tag indexes: {total_tags}")

    if google:
        for g in google:
            print(f"\n  ✓ 06.07.2021 google: id={g[0]}, amount={g[1]} cents = {g[1]/100:.2f} RUB, comment={g[2]}")
    else:
        print(f"\n  ✗ 06.07.2021 google row not found (target_time={target})")

    db.close()
    print("\nDone!")


if __name__ == '__main__':
    main()
