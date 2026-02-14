import { describe, it, expect, beforeEach, jest } from '@jest/globals';

// ─── CJS-style mocks (hoisted by jest) ──────────────────────────────────────

// Mock vue
jest.mock('vue', () => ({
    ref: (val: any) => ({ value: val }),
    computed: (fn: any) => ({ get value() { return fn(); } })
}));

// Mock @/locales/helpers.ts
const mockTt = jest.fn<(key: string) => string>().mockImplementation((key: string) => key);
jest.mock('@/locales/helpers.ts', () => ({
    useI18n: () => ({ tt: mockTt })
}));

// Mock @/core/datetime.ts
jest.mock('@/core/datetime.ts', () => ({
    DateRange: {
        All: { type: 0 },
        ThisWeek: { type: 5 },
        LastWeek: { type: 6 },
        ThisMonth: { type: 7 },
        LastMonth: { type: 8 },
        ThisQuarter: { type: 13 },
        ThisYear: { type: 9 },
        LastYear: { type: 10 }
    }
}));

// Mock @/lib/datetime.ts
const mockParseDateTimeFromUnixTime = jest.fn<(unixTime: number) => any>();
const mockGetUnixTimeBeforeUnixTime = jest.fn<(unixTime: number, amount: number, unit: string) => number>();
const mockGetUnixTimeAfterUnixTime = jest.fn<(unixTime: number, amount: number, unit: string) => number>();
const mockGetYearFirstUnixTime = jest.fn<(year: number) => number>();
const mockGetYearLastUnixTime = jest.fn<(year: number) => number>();
const mockGetQuarterFirstUnixTime = jest.fn<(yq: { year: number; quarter: number }) => number>();
const mockGetQuarterLastUnixTime = jest.fn<(yq: { year: number; quarter: number }) => number>();

jest.mock('@/lib/datetime.ts', () => ({
    parseDateTimeFromUnixTime: mockParseDateTimeFromUnixTime,
    getUnixTimeBeforeUnixTime: mockGetUnixTimeBeforeUnixTime,
    getUnixTimeAfterUnixTime: mockGetUnixTimeAfterUnixTime,
    getYearFirstUnixTime: mockGetYearFirstUnixTime,
    getYearLastUnixTime: mockGetYearLastUnixTime,
    getQuarterFirstUnixTime: mockGetQuarterFirstUnixTime,
    getQuarterLastUnixTime: mockGetQuarterLastUnixTime
}));

import { usePeriodNavigation } from '../usePeriodNavigation.ts';

// ─── Helpers ─────────────────────────────────────────────────────────────────

function makeFakeDateTime(year: number, month: number, day: number) {
    return {
        toGregorianCalendarYearMonthDay: () => ({ year, month, day })
    };
}

function computedVal<T>(getter: () => T) {
    return { get value() { return getter(); } } as any;
}

// ─── Tests ───────────────────────────────────────────────────────────────────

describe('usePeriodNavigation', () => {
    let dateTypeVal: number;
    let minTimeVal: number;
    let maxTimeVal: number;
    let onDateRangeChange: jest.Mock;

    function createNav() {
        return usePeriodNavigation({
            dateType: computedVal(() => dateTypeVal),
            minTime: computedVal(() => minTimeVal),
            maxTime: computedVal(() => maxTimeVal),
            onDateRangeChange
        });
    }

    beforeEach(() => {
        dateTypeVal = 0;
        minTimeVal = 0;
        maxTimeVal = 0;
        onDateRangeChange = jest.fn();
        mockTt.mockImplementation((key: string) => key);
        mockParseDateTimeFromUnixTime.mockReset();
        mockGetUnixTimeBeforeUnixTime.mockReset();
        mockGetUnixTimeAfterUnixTime.mockReset();
        mockGetYearFirstUnixTime.mockReset();
        mockGetYearLastUnixTime.mockReset();
        mockGetQuarterFirstUnixTime.mockReset();
        mockGetQuarterLastUnixTime.mockReset();
    });

    // ─── Return shape ────────────────────────────────────────────────────────

    it('should return navigationMode, periodLabel, navigatePeriod, resetNavigationMode, isAllTime', () => {
        const nav = createNav();
        expect(nav.navigationMode).toBeDefined();
        expect(nav.periodLabel).toBeDefined();
        expect(typeof nav.navigatePeriod).toBe('function');
        expect(typeof nav.resetNavigationMode).toBe('function');
        expect(nav.isAllTime).toBeDefined();
    });

    // ─── isAllTime ───────────────────────────────────────────────────────────

    it('should return isAllTime=true when dateType is All', () => {
        dateTypeVal = 0;
        const nav = createNav();
        expect(nav.isAllTime.value).toBe(true);
    });

    it('should return isAllTime=false when dateType is not All', () => {
        dateTypeVal = 7;
        const nav = createNav();
        expect(nav.isAllTime.value).toBe(false);
    });

    // ─── periodLabel ─────────────────────────────────────────────────────────

    it('should return "All time" label when dateType is All', () => {
        dateTypeVal = 0;
        const nav = createNav();
        expect(nav.periodLabel.value).toBe('All time');
        expect(mockTt).toHaveBeenCalledWith('All time');
    });

    it('should return month label when dateType is ThisMonth', () => {
        dateTypeVal = 7;
        minTimeVal = 1000000;
        mockParseDateTimeFromUnixTime.mockReturnValue(makeFakeDateTime(2025, 3, 1));

        const nav = createNav();
        expect(nav.periodLabel.value).toBe('month_standalone_3 2025');
    });

    it('should return month label when dateType is LastMonth', () => {
        dateTypeVal = 8;
        minTimeVal = 1000000;
        mockParseDateTimeFromUnixTime.mockReturnValue(makeFakeDateTime(2025, 2, 1));

        const nav = createNav();
        expect(nav.periodLabel.value).toBe('month_standalone_2 2025');
    });

    it('should return year label when dateType is ThisYear', () => {
        dateTypeVal = 9;
        minTimeVal = 1000000;
        mockParseDateTimeFromUnixTime.mockReturnValue(makeFakeDateTime(2025, 1, 1));

        const nav = createNav();
        expect(nav.periodLabel.value).toBe('2025');
    });

    it('should return year label when dateType is LastYear', () => {
        dateTypeVal = 10;
        minTimeVal = 1000000;
        mockParseDateTimeFromUnixTime.mockReturnValue(makeFakeDateTime(2024, 1, 1));

        const nav = createNav();
        expect(nav.periodLabel.value).toBe('2024');
    });

    it('should return quarter label when dateType is ThisQuarter', () => {
        dateTypeVal = 13;
        minTimeVal = 1000000;
        mockParseDateTimeFromUnixTime.mockReturnValue(makeFakeDateTime(2025, 7, 1));

        const nav = createNav();
        expect(nav.periodLabel.value).toBe('Q3 2025');
    });

    it('should return date range label when dateType has no recognized mode', () => {
        dateTypeVal = 3;
        minTimeVal = 1000000;
        maxTimeVal = 2000000;
        mockParseDateTimeFromUnixTime
            .mockReturnValueOnce(makeFakeDateTime(2025, 1, 15))
            .mockReturnValueOnce(makeFakeDateTime(2025, 1, 22));
        mockTt.mockImplementation((key: string) => {
            if (key === 'month_short_1') return 'Jan';
            return key;
        });

        const nav = createNav();
        expect(nav.periodLabel.value).toBe('15 Jan 2025 – 22 Jan 2025');
    });

    it('should return "Custom Date" when minTime and maxTime are both 0', () => {
        dateTypeVal = 255;
        minTimeVal = 0;
        maxTimeVal = 0;

        const nav = createNav();
        expect(nav.periodLabel.value).toBe('Custom Date');
    });

    it('should pad single-digit days in date range label', () => {
        dateTypeVal = 3;
        minTimeVal = 1000000;
        maxTimeVal = 2000000;
        mockParseDateTimeFromUnixTime
            .mockReturnValueOnce(makeFakeDateTime(2025, 5, 3))
            .mockReturnValueOnce(makeFakeDateTime(2025, 5, 9));
        mockTt.mockImplementation((key: string) => {
            if (key === 'month_short_5') return 'May';
            return key;
        });

        const nav = createNav();
        expect(nav.periodLabel.value).toBe('03 May 2025 – 09 May 2025');
    });

    // ─── periodLabel with navigationMode override ────────────────────────────

    it('should use navigationMode over dateType inference for periodLabel', () => {
        dateTypeVal = 3;
        minTimeVal = 1000000;
        mockParseDateTimeFromUnixTime.mockReturnValue(makeFakeDateTime(2025, 6, 1));

        const nav = createNav();
        nav.navigationMode.value = 'month';
        expect(nav.periodLabel.value).toBe('month_standalone_6 2025');
    });

    // ─── resetNavigationMode ─────────────────────────────────────────────────

    it('should reset navigationMode to empty string', () => {
        const nav = createNav();
        nav.navigationMode.value = 'month';
        nav.resetNavigationMode();
        expect(nav.navigationMode.value).toBe('');
    });

    // ─── navigatePeriod — guard clauses ──────────────────────────────────────

    it('should not navigate when dateType is All', () => {
        dateTypeVal = 0;
        minTimeVal = 1000;
        maxTimeVal = 2000;

        const nav = createNav();
        nav.navigatePeriod(1);
        expect(onDateRangeChange).not.toHaveBeenCalled();
    });

    it('should not navigate when minTime is 0', () => {
        dateTypeVal = 7;
        minTimeVal = 0;
        maxTimeVal = 2000;

        const nav = createNav();
        nav.navigatePeriod(1);
        expect(onDateRangeChange).not.toHaveBeenCalled();
    });

    it('should not navigate when maxTime is 0', () => {
        dateTypeVal = 7;
        minTimeVal = 1000;
        maxTimeVal = 0;

        const nav = createNav();
        nav.navigatePeriod(1);
        expect(onDateRangeChange).not.toHaveBeenCalled();
    });

    // ─── navigatePeriod — week mode ──────────────────────────────────────────

    it('should navigate forward by week when dateType is ThisWeek', () => {
        dateTypeVal = 5;
        minTimeVal = 1000;
        maxTimeVal = 2000;
        mockParseDateTimeFromUnixTime.mockReturnValue(makeFakeDateTime(2025, 1, 6));
        mockGetUnixTimeAfterUnixTime
            .mockReturnValueOnce(3000)
            .mockReturnValueOnce(4000);

        const nav = createNav();
        nav.navigatePeriod(1);

        expect(mockGetUnixTimeAfterUnixTime).toHaveBeenCalledWith(1000, 7, 'days');
        expect(mockGetUnixTimeAfterUnixTime).toHaveBeenCalledWith(2000, 7, 'days');
        expect(onDateRangeChange).toHaveBeenCalledWith(3000, 4000);
        expect(nav.navigationMode.value).toBe('week');
    });

    it('should navigate backward by week', () => {
        dateTypeVal = 5;
        minTimeVal = 1000;
        maxTimeVal = 2000;
        mockParseDateTimeFromUnixTime.mockReturnValue(makeFakeDateTime(2025, 1, 6));
        mockGetUnixTimeBeforeUnixTime
            .mockReturnValueOnce(500)
            .mockReturnValueOnce(1500);

        const nav = createNav();
        nav.navigatePeriod(-1);

        expect(mockGetUnixTimeBeforeUnixTime).toHaveBeenCalledWith(1000, 7, 'days');
        expect(mockGetUnixTimeBeforeUnixTime).toHaveBeenCalledWith(2000, 7, 'days');
        expect(onDateRangeChange).toHaveBeenCalledWith(500, 1500);
    });

    // ─── navigatePeriod — month mode ─────────────────────────────────────────

    it('should navigate forward by month when dateType is ThisMonth', () => {
        dateTypeVal = 7;
        minTimeVal = 1000;
        maxTimeVal = 2000;
        mockParseDateTimeFromUnixTime.mockReturnValue(makeFakeDateTime(2025, 3, 1));
        mockGetYearFirstUnixTime.mockReturnValue(100);
        mockGetUnixTimeAfterUnixTime
            .mockReturnValueOnce(5000)
            .mockReturnValueOnce(6001);

        const nav = createNav();
        nav.navigatePeriod(1);

        expect(mockGetYearFirstUnixTime).toHaveBeenCalledWith(2025);
        expect(mockGetUnixTimeAfterUnixTime).toHaveBeenCalledWith(100, 3, 'months');
        expect(mockGetUnixTimeAfterUnixTime).toHaveBeenCalledWith(5000, 1, 'months');
        expect(onDateRangeChange).toHaveBeenCalledWith(5000, 6000);
        expect(nav.navigationMode.value).toBe('month');
    });

    it('should wrap month forward from December to January of next year', () => {
        dateTypeVal = 7;
        minTimeVal = 1000;
        maxTimeVal = 2000;
        mockParseDateTimeFromUnixTime.mockReturnValue(makeFakeDateTime(2025, 12, 1));
        mockGetYearFirstUnixTime.mockReturnValue(200);
        mockGetUnixTimeAfterUnixTime
            .mockReturnValueOnce(7000)
            .mockReturnValueOnce(8001);

        const nav = createNav();
        nav.navigatePeriod(1);

        expect(mockGetYearFirstUnixTime).toHaveBeenCalledWith(2026);
        expect(mockGetUnixTimeAfterUnixTime).toHaveBeenCalledWith(200, 0, 'months');
        expect(onDateRangeChange).toHaveBeenCalledWith(7000, 8000);
    });

    it('should wrap month backward from January to December of previous year', () => {
        dateTypeVal = 7;
        minTimeVal = 1000;
        maxTimeVal = 2000;
        mockParseDateTimeFromUnixTime.mockReturnValue(makeFakeDateTime(2025, 1, 1));
        mockGetYearFirstUnixTime.mockReturnValue(300);
        mockGetUnixTimeAfterUnixTime
            .mockReturnValueOnce(9000)
            .mockReturnValueOnce(10001);

        const nav = createNav();
        nav.navigatePeriod(-1);

        expect(mockGetYearFirstUnixTime).toHaveBeenCalledWith(2024);
        expect(mockGetUnixTimeAfterUnixTime).toHaveBeenCalledWith(300, 11, 'months');
        expect(onDateRangeChange).toHaveBeenCalledWith(9000, 10000);
    });

    // ─── navigatePeriod — quarter mode ───────────────────────────────────────

    it('should navigate forward by quarter', () => {
        dateTypeVal = 13;
        minTimeVal = 1000;
        maxTimeVal = 2000;
        mockParseDateTimeFromUnixTime.mockReturnValue(makeFakeDateTime(2025, 4, 1));
        mockGetQuarterFirstUnixTime.mockReturnValue(11000);
        mockGetQuarterLastUnixTime.mockReturnValue(12000);

        const nav = createNav();
        nav.navigatePeriod(1);

        expect(mockGetQuarterFirstUnixTime).toHaveBeenCalledWith({ year: 2025, quarter: 3 });
        expect(mockGetQuarterLastUnixTime).toHaveBeenCalledWith({ year: 2025, quarter: 3 });
        expect(onDateRangeChange).toHaveBeenCalledWith(11000, 12000);
        expect(nav.navigationMode.value).toBe('quarter');
    });

    it('should wrap quarter forward from Q4 to Q1 of next year', () => {
        dateTypeVal = 13;
        minTimeVal = 1000;
        maxTimeVal = 2000;
        mockParseDateTimeFromUnixTime.mockReturnValue(makeFakeDateTime(2025, 10, 1));
        mockGetQuarterFirstUnixTime.mockReturnValue(13000);
        mockGetQuarterLastUnixTime.mockReturnValue(14000);

        const nav = createNav();
        nav.navigatePeriod(1);

        expect(mockGetQuarterFirstUnixTime).toHaveBeenCalledWith({ year: 2026, quarter: 1 });
        expect(onDateRangeChange).toHaveBeenCalledWith(13000, 14000);
    });

    it('should wrap quarter backward from Q1 to Q4 of previous year', () => {
        dateTypeVal = 13;
        minTimeVal = 1000;
        maxTimeVal = 2000;
        mockParseDateTimeFromUnixTime.mockReturnValue(makeFakeDateTime(2025, 2, 1));
        mockGetQuarterFirstUnixTime.mockReturnValue(15000);
        mockGetQuarterLastUnixTime.mockReturnValue(16000);

        const nav = createNav();
        nav.navigatePeriod(-1);

        expect(mockGetQuarterFirstUnixTime).toHaveBeenCalledWith({ year: 2024, quarter: 4 });
        expect(onDateRangeChange).toHaveBeenCalledWith(15000, 16000);
    });

    // ─── navigatePeriod — year mode ──────────────────────────────────────────

    it('should navigate forward by year', () => {
        dateTypeVal = 9;
        minTimeVal = 1000;
        maxTimeVal = 2000;
        mockParseDateTimeFromUnixTime.mockReturnValue(makeFakeDateTime(2025, 1, 1));
        mockGetYearFirstUnixTime.mockReturnValue(20000);
        mockGetYearLastUnixTime.mockReturnValue(21000);

        const nav = createNav();
        nav.navigatePeriod(1);

        expect(mockGetYearFirstUnixTime).toHaveBeenCalledWith(2026);
        expect(mockGetYearLastUnixTime).toHaveBeenCalledWith(2026);
        expect(onDateRangeChange).toHaveBeenCalledWith(20000, 21000);
        expect(nav.navigationMode.value).toBe('year');
    });

    it('should navigate backward by year', () => {
        dateTypeVal = 10;
        minTimeVal = 1000;
        maxTimeVal = 2000;
        mockParseDateTimeFromUnixTime.mockReturnValue(makeFakeDateTime(2024, 1, 1));
        mockGetYearFirstUnixTime.mockReturnValue(22000);
        mockGetYearLastUnixTime.mockReturnValue(23000);

        const nav = createNav();
        nav.navigatePeriod(-1);

        expect(mockGetYearFirstUnixTime).toHaveBeenCalledWith(2023);
        expect(mockGetYearLastUnixTime).toHaveBeenCalledWith(2023);
        expect(onDateRangeChange).toHaveBeenCalledWith(22000, 23000);
    });

    // ─── navigatePeriod — fallback (no recognized mode) ──────────────────────

    it('should use duration-based fallback when no mode is recognized', () => {
        dateTypeVal = 3;
        minTimeVal = 1000;
        maxTimeVal = 2000;
        mockParseDateTimeFromUnixTime.mockReturnValue(makeFakeDateTime(2025, 1, 1));

        const nav = createNav();
        nav.navigatePeriod(1);

        expect(onDateRangeChange).toHaveBeenCalledWith(2001, 3001);
    });

    it('should use duration-based fallback backward', () => {
        dateTypeVal = 3;
        minTimeVal = 3000;
        maxTimeVal = 5000;
        mockParseDateTimeFromUnixTime.mockReturnValue(makeFakeDateTime(2025, 1, 1));

        const nav = createNav();
        nav.navigatePeriod(-1);

        expect(onDateRangeChange).toHaveBeenCalledWith(999, 2999);
    });

    // ─── navigatePeriod with explicit navigationMode ─────────────────────────

    it('should use navigationMode over dateType inference for navigation', () => {
        dateTypeVal = 3;
        minTimeVal = 1000;
        maxTimeVal = 2000;
        mockParseDateTimeFromUnixTime.mockReturnValue(makeFakeDateTime(2025, 6, 1));
        mockGetYearFirstUnixTime.mockReturnValue(100);
        mockGetUnixTimeAfterUnixTime
            .mockReturnValueOnce(5000)
            .mockReturnValueOnce(6001);

        const nav = createNav();
        nav.navigationMode.value = 'month';
        nav.navigatePeriod(1);

        expect(mockGetYearFirstUnixTime).toHaveBeenCalled();
        expect(onDateRangeChange).toHaveBeenCalledWith(5000, 6000);
    });
});
