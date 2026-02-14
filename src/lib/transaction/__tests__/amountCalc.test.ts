import { describe, it, expect } from '@jest/globals';
import { calculateSuitableDestinationAmount } from '../amountCalc.ts';

describe('calculateSuitableDestinationAmount', () => {
    it('should return newSourceAmount when currencies are the same', () => {
        const result = calculateSuitableDestinationAmount(100, 200, 100, 'USD', 'USD');
        expect(result).toBe(200);
    });

    it('should return null when oldSourceAmount is zero', () => {
        const result = calculateSuitableDestinationAmount(0, 200, 100, 'USD', 'EUR');
        expect(result).toBeNull();
    });

    it('should return null when currentDestAmount is zero', () => {
        const result = calculateSuitableDestinationAmount(100, 200, 0, 'USD', 'EUR');
        expect(result).toBeNull();
    });

    it('should maintain exchange rate when different currencies', () => {
        // Old: 100 USD -> 90 EUR (rate = 0.9)
        // New: 200 USD -> should be 180 EUR
        const result = calculateSuitableDestinationAmount(100, 200, 90, 'USD', 'EUR');
        expect(result).toBe(180);
    });

    it('should round to nearest integer', () => {
        // Old: 300 USD -> 250 EUR (rate = 0.8333...)
        // New: 100 USD -> 83.33 -> rounds to 83
        const result = calculateSuitableDestinationAmount(300, 100, 250, 'USD', 'EUR');
        expect(result).toBe(83);
    });

    it('should handle rate > 1', () => {
        // Old: 100 USD -> 150 JPY (rate = 1.5)
        // New: 200 USD -> 300 JPY
        const result = calculateSuitableDestinationAmount(100, 200, 150, 'USD', 'JPY');
        expect(result).toBe(300);
    });

    it('should handle very small amounts', () => {
        const result = calculateSuitableDestinationAmount(1, 2, 1, 'USD', 'EUR');
        expect(result).toBe(2);
    });

    it('should handle equal source and dest initially', () => {
        const result = calculateSuitableDestinationAmount(100, 50, 100, 'USD', 'EUR');
        expect(result).toBe(50); // rate = 1.0
    });
});
