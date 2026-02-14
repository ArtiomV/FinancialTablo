import { describe, it, expect } from '@jest/globals';
import { buildTransactionListPageParams } from '../filterParams.ts';
import type { TransactionListFilter } from '@/stores/transaction.ts';
import { DateRange } from '@/core/datetime.ts';

function makeFilter(overrides: Partial<TransactionListFilter> = {}): TransactionListFilter {
    return {
        dateType: DateRange.ThisMonth.type,
        maxTime: 1735689600,
        minTime: 1733011200,
        type: 0,
        categoryIds: '',
        accountIds: '',
        tagFilter: '',
        amountFilter: '',
        keyword: '',
        ...overrides
    };
}

describe('buildTransactionListPageParams', () => {
    it('should return empty string for default filter with list page type', () => {
        // pageType 1 = List (default), dateType All should produce empty
        const filter = makeFilter({ dateType: DateRange.All.type, maxTime: 0, minTime: 0 });
        const result = buildTransactionListPageParams(filter, 0); // 0 = List
        expect(result).toBe('');
    });

    it('should include dateType when not All', () => {
        const filter = makeFilter();
        const result = buildTransactionListPageParams(filter, 1);
        expect(result).toContain(`dateType=${DateRange.ThisMonth.type}`);
    });

    it('should include minTime and maxTime', () => {
        const filter = makeFilter({ minTime: 1000, maxTime: 2000 });
        const result = buildTransactionListPageParams(filter, 1);
        expect(result).toContain('minTime=1000');
        expect(result).toContain('maxTime=2000');
    });

    it('should include type when non-zero', () => {
        const filter = makeFilter({ type: 2 });
        const result = buildTransactionListPageParams(filter, 1);
        expect(result).toContain('type=2');
    });

    it('should not include type when zero', () => {
        const filter = makeFilter({ type: 0 });
        const result = buildTransactionListPageParams(filter, 1);
        expect(result).not.toContain('type=');
    });

    it('should encode categoryIds', () => {
        const filter = makeFilter({ categoryIds: '123,456' });
        const result = buildTransactionListPageParams(filter, 1);
        expect(result).toContain('categoryIds=123%2C456');
    });

    it('should encode accountIds', () => {
        const filter = makeFilter({ accountIds: '789' });
        const result = buildTransactionListPageParams(filter, 1);
        expect(result).toContain('accountIds=789');
    });

    it('should include keyword when non-empty', () => {
        const filter = makeFilter({ keyword: 'salary' });
        const result = buildTransactionListPageParams(filter, 1);
        expect(result).toContain('keyword=salary');
    });

    it('should include amountFilter', () => {
        const filter = makeFilter({ amountFilter: 'gte:100' });
        const result = buildTransactionListPageParams(filter, 1);
        expect(result).toContain('amountFilter=gte%3A100');
    });

    it('should include pageType when not List', () => {
        const filter = makeFilter();
        const result = buildTransactionListPageParams(filter, 2); // Calendar
        expect(result).toContain('pageType=2');
    });

    it('should join params with &', () => {
        const filter = makeFilter({ type: 2, keyword: 'test' });
        const result = buildTransactionListPageParams(filter, 1);
        const parts = result.split('&');
        expect(parts.length).toBeGreaterThan(1);
    });
});
