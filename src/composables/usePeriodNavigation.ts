/**
 * Composable for period-based navigation (week, month, quarter, year).
 *
 * Encapsulates the logic for navigating between time periods
 * (previous/next week, month, quarter, year), maintaining navigation
 * mode, and generating human-readable period labels.
 *
 * Extracted from transaction ListPage for reuse across reports,
 * statistics, and other period-based views.
 *
 * Usage:
 * ```ts
 * const { navigationMode, periodLabel, navigatePeriod } = usePeriodNavigation({
 *   dateType: computed(() => query.dateType),
 *   minTime: computed(() => query.minTime),
 *   maxTime: computed(() => query.maxTime),
 *   onDateRangeChange: (min, max) => updateFilter(min, max)
 * });
 * ```
 */

import { ref, computed, type Ref, type ComputedRef } from 'vue';
import { useI18n } from '@/locales/helpers.ts';
import { DateRange } from '@/core/datetime.ts';
import {
    parseDateTimeFromUnixTime,
    getUnixTimeBeforeUnixTime,
    getUnixTimeAfterUnixTime,
    getYearFirstUnixTime,
    getYearLastUnixTime,
    getQuarterFirstUnixTime,
    getQuarterLastUnixTime,
    getYearMonthFirstUnixTime,
    getYearMonthLastUnixTime
} from '@/lib/datetime.ts';

export interface UsePeriodNavigationOptions {
    /** Current date range type */
    dateType: ComputedRef<number>;
    /** Current period minimum unix time */
    minTime: ComputedRef<number>;
    /** Current period maximum unix time */
    maxTime: ComputedRef<number>;
    /** Callback when date range should change */
    onDateRangeChange: (minTime: number, maxTime: number) => void;
}

export interface UsePeriodNavigationReturn {
    /** Current navigation mode ('week', 'month', 'quarter', 'year', '') */
    navigationMode: Ref<string>;
    /** Human-readable period label */
    periodLabel: ComputedRef<string>;
    /** Navigate forward/backward */
    navigatePeriod: (direction: number) => void;
    /** Reset navigation mode (e.g., when user picks from menu) */
    resetNavigationMode: () => void;
    /** Whether the current period is 'all time' */
    isAllTime: ComputedRef<boolean>;
}

export function usePeriodNavigation(options: UsePeriodNavigationOptions): UsePeriodNavigationReturn {
    const { tt } = useI18n();

    const navigationMode = ref<string>('');

    const isAllTime = computed<boolean>(() => options.dateType.value === DateRange.All.type);

    function formatShortDateWithMonthName(unixTime: number): string {
        const dt = parseDateTimeFromUnixTime(unixTime);
        const ymd = dt.toGregorianCalendarYearMonthDay();
        const day = ymd.day < 10 ? '0' + ymd.day : String(ymd.day);
        const monthShort = tt('month_short_' + ymd.month);
        return `${day} ${monthShort} ${ymd.year}`;
    }

    const periodLabel = computed<string>(() => {
        const dt = options.dateType.value;
        const minTime = options.minTime.value;
        const maxTime = options.maxTime.value;

        if (dt === DateRange.All.type) {
            return tt('All time');
        }

        const mode = navigationMode.value ||
            ((dt === DateRange.ThisMonth.type || dt === DateRange.LastMonth.type) ? 'month' :
            (dt === DateRange.ThisYear.type || dt === DateRange.LastYear.type) ? 'year' :
            (dt === DateRange.ThisQuarter.type) ? 'quarter' : '');

        if (mode === 'month' && minTime) {
            const minDateTime = parseDateTimeFromUnixTime(minTime);
            const ymd = minDateTime.toGregorianCalendarYearMonthDay();
            return tt('month_standalone_' + ymd.month) + ' ' + ymd.year;
        }

        if (mode === 'year' && minTime) {
            const minDateTime = parseDateTimeFromUnixTime(minTime);
            const ymd = minDateTime.toGregorianCalendarYearMonthDay();
            return String(ymd.year);
        }

        if (mode === 'quarter' && minTime) {
            const minDateTime = parseDateTimeFromUnixTime(minTime);
            const ymd = minDateTime.toGregorianCalendarYearMonthDay();
            const q = Math.ceil(ymd.month / 3);
            return `Q${q} ${ymd.year}`;
        }

        if (minTime && maxTime) {
            return `${formatShortDateWithMonthName(minTime)} – ${formatShortDateWithMonthName(maxTime)}`;
        }

        return tt('Custom Date');
    });

    function navigatePeriod(direction: number): void {
        const dt = options.dateType.value;
        const currentMin = options.minTime.value;
        const currentMax = options.maxTime.value;

        if (!currentMin || !currentMax || dt === DateRange.All.type) {
            return;
        }

        let newMin: number;
        let newMax: number;

        const minDt = parseDateTimeFromUnixTime(currentMin);
        const ymd = minDt.toGregorianCalendarYearMonthDay();

        const mode = navigationMode.value ||
            ((dt === DateRange.ThisWeek.type || dt === DateRange.LastWeek.type) ? 'week' :
            (dt === DateRange.ThisMonth.type || dt === DateRange.LastMonth.type) ? 'month' :
            (dt === DateRange.ThisQuarter.type) ? 'quarter' :
            (dt === DateRange.ThisYear.type || dt === DateRange.LastYear.type) ? 'year' : '');

        if (mode === 'week') {
            if (direction > 0) {
                newMin = getUnixTimeAfterUnixTime(currentMin, 7, 'days');
                newMax = getUnixTimeAfterUnixTime(currentMax, 7, 'days');
            } else {
                newMin = getUnixTimeBeforeUnixTime(currentMin, 7, 'days');
                newMax = getUnixTimeBeforeUnixTime(currentMax, 7, 'days');
            }
            navigationMode.value = 'week';
        } else if (mode === 'month') {
            let targetMonth = ymd.month + direction;
            let targetYear = ymd.year;
            if (targetMonth > 12) { targetMonth -= 12; targetYear++; }
            if (targetMonth < 1) { targetMonth += 12; targetYear--; }
            newMin = getYearMonthFirstUnixTime({ year: targetYear, month1base: targetMonth });
            newMax = getYearMonthLastUnixTime({ year: targetYear, month1base: targetMonth });
            navigationMode.value = 'month';
        } else if (mode === 'quarter') {
            const currentQuarter = Math.ceil(ymd.month / 3);
            let targetQuarter = currentQuarter + direction;
            let targetYear = ymd.year;
            if (targetQuarter > 4) { targetQuarter = 1; targetYear++; }
            if (targetQuarter < 1) { targetQuarter = 4; targetYear--; }
            newMin = getQuarterFirstUnixTime({ year: targetYear, quarter: targetQuarter });
            newMax = getQuarterLastUnixTime({ year: targetYear, quarter: targetQuarter });
            navigationMode.value = 'quarter';
        } else if (mode === 'year') {
            const targetYear = ymd.year + direction;
            newMin = getYearFirstUnixTime(targetYear);
            newMax = getYearLastUnixTime(targetYear);
            navigationMode.value = 'year';
        } else {
            const duration = currentMax - currentMin;
            if (direction > 0) {
                newMin = currentMax + 1;
                newMax = newMin + duration;
            } else {
                newMax = currentMin - 1;
                newMin = newMax - duration;
            }
        }

        options.onDateRangeChange(newMin, newMax);
    }

    function resetNavigationMode(): void {
        navigationMode.value = '';
    }

    return {
        navigationMode,
        periodLabel,
        navigatePeriod,
        resetNavigationMode,
        isAllTime
    };
}
