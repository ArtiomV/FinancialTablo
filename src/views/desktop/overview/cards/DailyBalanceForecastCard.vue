<template>
    <div :class="{ 'disabled': disabled }">
        <div class="daily-balance-chart-container daily-balance-chart-overlay" v-if="loading && !hasAnyData">
            <div class="d-flex flex-column align-center justify-center w-100 h-100">
                <v-skeleton-loader class="w-100" type="image" :loading="true"></v-skeleton-loader>
            </div>
        </div>

        <div class="daily-balance-chart-container daily-balance-chart-overlay" v-else-if="!loading && !hasAnyData">
            <div class="d-flex flex-column align-center justify-center w-100 h-100">
                <span class="text-medium-emphasis">{{ tt('No data') }}</span>
            </div>
        </div>

        <v-chart autoresize class="daily-balance-chart-container"
                 :class="{ 'readonly': !hasAnyData }" :option="chartOptions"/>
    </div>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import type { CallbackDataParams } from 'echarts/types/dist/shared';

import { useI18n } from '@/locales/helpers.ts';

import { useSettingsStore } from '@/stores/setting.ts';
import { useUserStore } from '@/stores/user.ts';

import type { HiddenAmount } from '@/core/numeral.ts';
import { DISPLAY_HIDDEN_AMOUNT } from '@/consts/numeral.ts';

export interface DailyBalanceForecastDataItem {
    date: string;
    dateLabel: string;
    balance: number;
    isFuture: boolean;
    dailyIncome?: number;
    dailyExpense?: number;
}

const props = defineProps<{
    loading: boolean;
    data: DailyBalanceForecastDataItem[];
    disabled: boolean;
    isDarkMode?: boolean;
}>();

const {
    tt,
    formatAmountToLocalizedNumeralsWithCurrency
} = useI18n();

const settingsStore = useSettingsStore();
const userStore = useUserStore();

const showAmountInHomePage = computed<boolean>(() => settingsStore.appSettings.showAmountInHomePage);
const defaultCurrency = computed<string>(() => userStore.currentUserDefaultCurrency);

const hasAnyData = computed<boolean>(() => {
    return props.data && props.data.length > 0;
});

function getDisplayCurrency(value: number | HiddenAmount, currencyCode: string): string {
    return formatAmountToLocalizedNumeralsWithCurrency(value, currencyCode);
}

function formatBalanceForDisplay(balance: number): string {
    if (!showAmountInHomePage.value) {
        return getDisplayCurrency(DISPLAY_HIDDEN_AMOUNT, defaultCurrency.value);
    }
    return formatAmountToLocalizedNumeralsWithCurrency(balance, defaultCurrency.value);
}

// Segment: a contiguous run of points that are all positive or all negative
interface Segment {
    startIndex: number;
    endIndex: number;
    isNegative: boolean;
    isFuture: boolean; // true if ALL points in segment are future
    isMixed: boolean;  // true if segment spans past and future
}

const chartOptions = computed<object>(() => {
    const dayLabels: string[] = [];
    const dateLabels: string[] = [];
    const allBalances: number[] = [];
    let todayIndex = -1;

    if (props.data) {
        for (let i = 0; i < props.data.length; i++) {
            const item = props.data[i]!;
            dayLabels.push(item.date);
            dateLabels.push(item.dateLabel);
            allBalances.push(item.balance);

            if (!item.isFuture) {
                todayIndex = i;
            }
        }
    }

    const primaryColor = props.isDarkMode ? '#8C9EFF' : '#5C6BC0';
    const negativeColor = props.isDarkMode ? '#EF5350' : '#D32F2F';
    const negativeAreaColor = props.isDarkMode ? 'rgba(239, 83, 80, 0.15)' : 'rgba(211, 47, 47, 0.10)';

    // NO visualMap. Instead, split data into segments by sign.
    // Each segment becomes its own series with explicit color.
    // Adjacent segments share their boundary point for continuity.

    // Step 1: Find sign-change boundaries
    // A segment is a maximal run where all values have the same sign (>=0 or <0).
    // Boundary points are shared between adjacent segments.
    const segments: Segment[] = [];
    if (allBalances.length > 0) {
        let segStart = 0;
        let segNeg = allBalances[0]! < 0;

        for (let i = 1; i < allBalances.length; i++) {
            const curNeg = allBalances[i]! < 0;
            if (curNeg !== segNeg) {
                // Sign changed: close current segment at i-1, start new at i-1 (shared point)
                const allFuture = segStart > todayIndex;
                const allPast = (i - 1) <= todayIndex;
                segments.push({
                    startIndex: segStart,
                    endIndex: i - 1,
                    isNegative: segNeg,
                    isFuture: allFuture,
                    isMixed: !allFuture && !allPast
                });
                segStart = i - 1; // shared boundary point
                segNeg = curNeg;
            }
        }
        // Close last segment
        const allFuture = segStart > todayIndex;
        const allPast = (allBalances.length - 1) <= todayIndex;
        segments.push({
            startIndex: segStart,
            endIndex: allBalances.length - 1,
            isNegative: segNeg,
            isFuture: allFuture,
            isMixed: !allFuture && !allPast
        });
    }

    // Step 2: For each segment, build a series with [categoryName, value] data
    // Use explicit lineStyle.color and itemStyle.color (NO visualMap)
    const series: object[] = [];
    let firstSeriesIdx = -1; // index of first series (for markLine/markArea)

    for (let s = 0; s < segments.length; s++) {
        const seg = segments[s]!;
        const color = seg.isNegative ? negativeColor : primaryColor;

        // Determine line style: dashed for future, solid for past
        // If segment spans the today boundary, we need to split it further
        // For simplicity: if ANY point is future, check per-point

        // Build data array for this segment
        const segData: [string, number][] = [];
        for (let i = seg.startIndex; i <= seg.endIndex; i++) {
            segData.push([dayLabels[i]!, allBalances[i]!]);
        }

        // Determine if this segment is entirely past, entirely future, or mixed
        const entirelyPast = seg.endIndex <= todayIndex;
        const entirelyFuture = seg.startIndex > todayIndex;

        if (entirelyPast || entirelyFuture) {
            // Simple case: one series for the whole segment
            const isFuture = entirelyFuture;
            const lineType = isFuture ? 'dashed' : 'solid';

            const seriesObj: Record<string, unknown> = {
                type: 'line',
                name: s === 0 ? tt('Balance') : undefined, // only first gets the name for legend
                data: segData,
                smooth: true,
                showSymbol: true,
                symbol: 'circle',
                symbolSize: (_value: [string, number] | null, params: { dataIndex: number }) => {
                    if (!_value) return 0;
                    const globalIdx = seg.startIndex + (params.dataIndex ?? 0);
                    return globalIdx === todayIndex ? 8 : 4;
                },
                lineStyle: { color: color, width: 2.5, type: lineType },
                itemStyle: { color: color },
                areaStyle: {
                    color: seg.isNegative ? negativeAreaColor : (props.isDarkMode ? 'rgba(140, 158, 255, 0.15)' : 'rgba(92, 107, 192, 0.1)'),
                    origin: 0
                },
                z: isFuture ? 2 : 1
            };

            if (firstSeriesIdx === -1) {
                firstSeriesIdx = series.length;
            }

            series.push(seriesObj);
        } else {
            // Mixed segment: crosses today boundary. Split into past part and future part.
            // Past part: startIndex..todayIndex
            // Future part: todayIndex..endIndex (shared point at todayIndex)
            const pastData: [string, number][] = [];
            for (let i = seg.startIndex; i <= todayIndex; i++) {
                pastData.push([dayLabels[i]!, allBalances[i]!]);
            }
            const futureData: [string, number][] = [];
            for (let i = todayIndex; i <= seg.endIndex; i++) {
                futureData.push([dayLabels[i]!, allBalances[i]!]);
            }

            if (firstSeriesIdx === -1) {
                firstSeriesIdx = series.length;
            }

            // Past sub-series (solid)
            series.push({
                type: 'line',
                name: s === 0 ? tt('Balance') : undefined,
                data: pastData,
                smooth: true,
                showSymbol: true,
                symbol: 'circle',
                symbolSize: (_value: [string, number] | null, params: { dataIndex: number }) => {
                    if (!_value) return 0;
                    const globalIdx = seg.startIndex + (params.dataIndex ?? 0);
                    return globalIdx === todayIndex ? 8 : 4;
                },
                lineStyle: { color: color, width: 2.5, type: 'solid' },
                itemStyle: { color: color },
                areaStyle: {
                    color: seg.isNegative ? negativeAreaColor : (props.isDarkMode ? 'rgba(140, 158, 255, 0.15)' : 'rgba(92, 107, 192, 0.1)'),
                    origin: 0
                },
                z: 1
            });

            // Future sub-series (dashed)
            series.push({
                type: 'line',
                data: futureData,
                smooth: true,
                showSymbol: true,
                symbol: 'circle',
                symbolSize: 4,
                lineStyle: { color: color, width: 2.5, type: 'dashed' },
                itemStyle: { color: color },
                areaStyle: {
                    color: seg.isNegative ? negativeAreaColor : (props.isDarkMode ? 'rgba(140, 158, 255, 0.08)' : 'rgba(92, 107, 192, 0.06)'),
                    origin: 0
                },
                z: 2
            });
        }
    }

    // If no segments were created (no data), add an empty series
    if (series.length === 0) {
        series.push({
            type: 'line',
            data: [],
            smooth: true
        });
        firstSeriesIdx = 0;
    }

    // Add markLine (Today) and markArea (negative zones) to the first series
    if (firstSeriesIdx >= 0 && firstSeriesIdx < series.length) {
        const firstSeries = series[firstSeriesIdx] as Record<string, unknown>;

        if (todayIndex >= 0) {
            firstSeries['markLine'] = {
                silent: true,
                symbol: 'none',
                lineStyle: {
                    color: props.isDarkMode ? 'rgba(255,255,255,0.4)' : 'rgba(0,0,0,0.25)',
                    type: 'dashed',
                    width: 1
                },
                data: [
                    { xAxis: dayLabels[todayIndex] }
                ],
                label: {
                    show: true,
                    formatter: tt('Today'),
                    position: 'insideStartTop',
                    color: props.isDarkMode ? '#aaa' : '#666',
                    fontSize: 11
                }
            };
        }

        // Build markArea for negative zones
        const negativeAreas: [Record<string, unknown>, Record<string, unknown>][] = [];
        if (props.data) {
            let inNeg = false;
            let startIdx = 0;
            for (let i = 0; i < props.data.length; i++) {
                const item = props.data[i]!;
                if (item.balance < 0 && !inNeg) {
                    inNeg = true;
                    startIdx = i;
                } else if ((item.balance >= 0 || i === props.data.length - 1) && inNeg) {
                    const endIdx = item.balance < 0 ? i : i - 1;
                    negativeAreas.push([
                        { xAxis: dayLabels[startIdx] },
                        { xAxis: dayLabels[endIdx] }
                    ]);
                    inNeg = false;
                }
            }
        }

        if (negativeAreas.length > 0) {
            firstSeries['markArea'] = {
                silent: true,
                itemStyle: {
                    color: negativeAreaColor
                },
                data: negativeAreas
            };
        }
    }

    // Tooltip formatter: find the correct global index from any series
    const tooltipFormatter = (params: CallbackDataParams | CallbackDataParams[]) => {
        const paramArray = Array.isArray(params) ? params : [params];
        let bestParam: CallbackDataParams | null = null;
        for (const p of paramArray) {
            if (p.value !== null && p.value !== undefined && p.value !== '-') {
                // For [label, value] data, value is the pair array
                const val = Array.isArray(p.value) ? p.value[1] : p.value;
                if (val !== null && val !== undefined) {
                    bestParam = p;
                    break;
                }
            }
        }
        if (!bestParam) {
            return '';
        }

        // Extract the day label from the data point
        const pointValue = bestParam.value as [string, number];
        const dayLabel = Array.isArray(pointValue) ? pointValue[0] : bestParam.name;

        // Find global index by matching day label
        let globalIndex = 0;
        for (let i = 0; i < dayLabels.length; i++) {
            if (dayLabels[i] === dayLabel) {
                globalIndex = i;
                break;
            }
        }

        const dateLabelText = dateLabels[globalIndex] || bestParam.name || '';
        const balanceValue = allBalances[globalIndex]!;
        const balance = formatBalanceForDisplay(balanceValue);
        const isFuture = globalIndex > todayIndex;
        const isNeg = balanceValue < 0;
        const label = isFuture ? tt('Forecast') : tt('Balance');
        const color = isNeg ? negativeColor : primaryColor;
        const balanceHtml = isNeg
            ? `<strong style="color: ${negativeColor}">${balance}</strong>`
            : `<strong>${balance}</strong>`;

        let html = `<div><strong>${dateLabelText}</strong></div>` +
            `<div><span class="daily-balance-chart-tooltip-indicator" style="background-color: ${color}"></span> ${label}: ${balanceHtml}</div>`;

        // Show daily income/expense if available
        const dataItem = props.data && props.data[globalIndex];
        if (dataItem) {
            const incomeColor = '#4CAF50';
            const expenseColor = '#F44336';
            if (dataItem.dailyIncome && dataItem.dailyIncome > 0) {
                const incomeFormatted = formatBalanceForDisplay(dataItem.dailyIncome);
                html += `<div><span class="daily-balance-chart-tooltip-indicator" style="background-color: ${incomeColor}"></span> ${tt('Income')}: <strong style="color: ${incomeColor}">+${incomeFormatted}</strong></div>`;
            }
            if (dataItem.dailyExpense && dataItem.dailyExpense > 0) {
                const expenseFormatted = formatBalanceForDisplay(dataItem.dailyExpense);
                html += `<div><span class="daily-balance-chart-tooltip-indicator" style="background-color: ${expenseColor}"></span> ${tt('Expense')}: <strong style="color: ${expenseColor}">\u2013${expenseFormatted}</strong></div>`;
            }
        }

        return html;
    };

    return {
        tooltip: {
            trigger: 'axis',
            backgroundColor: props.isDarkMode ? '#333' : '#fff',
            borderColor: props.isDarkMode ? '#333' : '#fff',
            textStyle: {
                color: props.isDarkMode ? '#eee' : '#333'
            },
            formatter: tooltipFormatter
        },
        grid: {
            left: '60px',
            right: '20px',
            top: '15px',
            bottom: '30px'
        },
        legend: {
            show: false
        },
        xAxis: [
            {
                type: 'category',
                data: dayLabels,
                boundaryGap: false,
                axisLine: {
                    lineStyle: {
                        color: props.isDarkMode ? '#555' : '#ccc'
                    }
                },
                axisTick: {
                    show: false
                },
                axisLabel: {
                    color: props.isDarkMode ? '#aaa' : '#666',
                    interval: 'auto',
                    fontSize: 10,
                    rotate: 0
                }
            }
        ],
        yAxis: [
            {
                type: 'value',
                axisLabel: {
                    color: props.isDarkMode ? '#aaa' : '#666',
                    formatter: (value: number) => {
                        if (Math.abs(value) >= 100000000) {
                            return (value / 100000000).toFixed(0) + 'M';
                        } else if (Math.abs(value) >= 100000) {
                            return (value / 100000).toFixed(0) + 'K';
                        } else if (Math.abs(value) >= 100) {
                            return (value / 100).toFixed(0);
                        }
                        return String(value);
                    }
                },
                splitLine: {
                    lineStyle: {
                        color: props.isDarkMode ? 'rgba(255,255,255,0.06)' : 'rgba(0,0,0,0.06)'
                    }
                }
            }
        ],
        series: series
    };
});
</script>

<style>
.daily-balance-chart-container {
    width: 100%;
    height: 220px;
}

.daily-balance-chart-overlay {
    position: absolute !important;
    z-index: 10;
}

.daily-balance-chart-tooltip-indicator {
    display: inline-block;
    width: 10px;
    height: 10px;
    border-radius: 10px;
    margin-right: 4px;
}
</style>
