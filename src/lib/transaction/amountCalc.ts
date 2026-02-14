/**
 * Transaction amount calculation utilities.
 *
 * Pure functions for calculating display amounts, suitable exchange amounts,
 * and other numeric transformations on transactions. These are extracted
 * from the transaction store for reusability and testability.
 */

import { getCurrencyFraction } from '@/lib/currency.ts';

/**
 * Calculate the suitable destination amount for a transfer transaction
 * when the source amount changes, maintaining the exchange rate ratio.
 *
 * @param oldSourceAmount - Previous source amount
 * @param newSourceAmount - New source amount
 * @param currentDestAmount - Current destination amount
 * @param sourceCurrency - Source account currency code
 * @param destCurrency - Destination account currency code
 * @returns The new destination amount, or null if no change is needed
 */
export function calculateSuitableDestinationAmount(
    oldSourceAmount: number,
    newSourceAmount: number,
    currentDestAmount: number,
    sourceCurrency: string,
    destCurrency: string
): number | null {
    if (sourceCurrency === destCurrency) {
        // Same currency: destination always equals source
        return newSourceAmount;
    }

    if (oldSourceAmount === 0 || currentDestAmount === 0) {
        return null; // Can't calculate rate
    }

    // Maintain the exchange rate ratio
    const rate = currentDestAmount / oldSourceAmount;
    return Math.round(newSourceAmount * rate);
}

/**
 * Get the number of decimal places for a currency.
 *
 * @param currencyCode - ISO currency code (e.g., 'USD', 'JPY')
 * @returns Number of decimal places (0, 2, or 3 typically)
 */
export function getCurrencyDecimalPlaces(currencyCode: string): number {
    const fraction = getCurrencyFraction(currencyCode) ?? 100;
    if (fraction <= 1) return 0;
    if (fraction <= 10) return 1;
    if (fraction <= 100) return 2;
    return 3;
}

/**
 * Round an amount to the appropriate precision for a currency.
 *
 * @param amount - Raw amount in minor units
 * @param currencyCode - ISO currency code
 * @returns Rounded amount
 */
export function roundAmountForCurrency(amount: number, currencyCode: string): number {
    const fraction = getCurrencyFraction(currencyCode) ?? 100;
    if (fraction <= 1) {
        return Math.round(amount / 100) * 100;
    }
    return Math.round(amount);
}
