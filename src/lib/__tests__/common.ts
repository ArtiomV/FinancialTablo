// Unit tests for common utility functions
import { describe, expect, test } from '@jest/globals';

import {
    isDefined,
    isObject,
    isArray,
    isString,
    isNumber,
    isInteger,
    isBoolean,
    isYearMonth,
    isEquals,
    isYearMonthEquals,
    isArray1SubsetOfArray2,
    isObjectEmpty,
    getNumberValue,
    sortNumbersArray,
    replaceAll,
    removeAll
} from '@/lib/common.ts';

// TYPE GUARDS

describe('isDefined', () => {
    test('returns false for null', () => {
        expect(isDefined(null)).toBe(false);
    });

    test('returns false for undefined', () => {
        expect(isDefined(undefined)).toBe(false);
    });

    test('returns true for 0', () => {
        expect(isDefined(0)).toBe(true);
    });

    test('returns true for empty string', () => {
        expect(isDefined("")).toBe(true);
    });

    test('returns true for empty object', () => {
        expect(isDefined({})).toBe(true);
    });
});

describe('isObject', () => {
    test('returns true for plain object', () => {
        expect(isObject({})).toBe(true);
    });

    test('returns false for array', () => {
        expect(isObject([])).toBe(false);
    });

    test('returns false for null', () => {
        expect(isObject(null)).toBe(false);
    });

    test('returns false for string', () => {
        expect(isObject("string")).toBe(false);
    });
});

describe('isArray', () => {
    test('returns true for array', () => {
        expect(isArray([])).toBe(true);
    });

    test('returns false for object', () => {
        expect(isArray({})).toBe(false);
    });

    test('returns false for null', () => {
        expect(isArray(null)).toBe(false);
    });
});

describe('isString', () => {
    test('returns true for empty string', () => {
        expect(isString("")).toBe(true);
    });

    test('returns true for non-empty string', () => {
        expect(isString("hello")).toBe(true);
    });

    test('returns false for number', () => {
        expect(isString(123)).toBe(false);
    });
});

describe('isNumber', () => {
    test('returns true for integer', () => {
        expect(isNumber(123)).toBe(true);
    });

    test('returns true for NaN', () => {
        expect(isNumber(NaN)).toBe(true);
    });

    test('returns false for numeric string', () => {
        expect(isNumber("123")).toBe(false);
    });
});

describe('isInteger', () => {
    test('returns true for integer', () => {
        expect(isInteger(1)).toBe(true);
    });

    test('returns false for float', () => {
        expect(isInteger(1.5)).toBe(false);
    });

    test('returns false for NaN', () => {
        expect(isInteger(NaN)).toBe(false);
    });
});

describe('isBoolean', () => {
    test('returns true for true', () => {
        expect(isBoolean(true)).toBe(true);
    });

    test('returns true for false', () => {
        expect(isBoolean(false)).toBe(true);
    });

    test('returns false for 0', () => {
        expect(isBoolean(0)).toBe(false);
    });
});

// isYearMonth

describe('isYearMonth', () => {
    test('returns true for valid year-month "2024-01"', () => {
        expect(isYearMonth("2024-01")).toBe(true);
    });

    test('returns true for "2024-13" (only checks parsability)', () => {
        expect(isYearMonth("2024-13")).toBe(true);
    });

    test('returns false for non-numeric string "abc"', () => {
        expect(isYearMonth("abc")).toBe(false);
    });

    test('returns false for year only "2024"', () => {
        expect(isYearMonth("2024")).toBe(false);
    });

    test('returns false for empty string', () => {
        expect(isYearMonth("")).toBe(false);
    });
});

// isEquals

describe('isEquals', () => {
    test('returns true for same primitives', () => {
        expect(isEquals(1, 1)).toBe(true);
        expect(isEquals("hello", "hello")).toBe(true);
        expect(isEquals(true, true)).toBe(true);
    });

    test('returns false for different primitives', () => {
        expect(isEquals(1, 2)).toBe(false);
        expect(isEquals("hello", "world")).toBe(false);
    });

    test('returns true for same arrays', () => {
        expect(isEquals([1, 2, 3], [1, 2, 3])).toBe(true);
    });

    test('returns false for different arrays', () => {
        expect(isEquals([1, 2, 3], [1, 2, 4])).toBe(false);
    });

    test('returns false for arrays with different length', () => {
        expect(isEquals([1, 2], [1, 2, 3])).toBe(false);
    });

    test('returns true for same objects', () => {
        expect(isEquals({ a: 1, b: 2 }, { a: 1, b: 2 })).toBe(true);
    });

    test('returns false for objects with different values', () => {
        expect(isEquals({ a: 1, b: 2 }, { a: 1, b: 3 })).toBe(false);
    });

    test('returns false for objects with different keys', () => {
        expect(isEquals({ a: 1, b: 2 }, { a: 1, c: 2 })).toBe(false);
    });

    test('works recursively with nested objects', () => {
        expect(isEquals({ a: { b: { c: 1 } } }, { a: { b: { c: 1 } } })).toBe(true);
        expect(isEquals({ a: { b: { c: 1 } } }, { a: { b: { c: 2 } } })).toBe(false);
    });
});

// isYearMonthEquals

describe('isYearMonthEquals', () => {
    test('returns true for "2024-01" and "2024-1" (parsed as same integers)', () => {
        expect(isYearMonthEquals("2024-01", "2024-1")).toBe(true);
    });

    test('returns false for "2024-01" and "2024-02"', () => {
        expect(isYearMonthEquals("2024-01", "2024-02")).toBe(false);
    });
});

// isArray1SubsetOfArray2

describe('isArray1SubsetOfArray2', () => {
    test('returns true when array1 is subset of array2', () => {
        expect(isArray1SubsetOfArray2([1, 2], [1, 2, 3])).toBe(true);
    });

    test('returns false when array1 contains elements not in array2', () => {
        expect(isArray1SubsetOfArray2([1, 4], [1, 2, 3])).toBe(false);
    });

    test('returns true for empty subset', () => {
        expect(isArray1SubsetOfArray2([], [1, 2, 3])).toBe(true);
    });
});

// isObjectEmpty

describe('isObjectEmpty', () => {
    test('returns true for empty object', () => {
        expect(isObjectEmpty({})).toBe(true);
    });

    test('returns false for non-empty object', () => {
        expect(isObjectEmpty({ a: 1 })).toBe(false);
    });
});

// getNumberValue

describe('getNumberValue', () => {
    test('parses string to number', () => {
        expect(getNumberValue("123", 0)).toBe(123);
    });

    test('returns number value directly', () => {
        expect(getNumberValue(456, 0)).toBe(456);
    });

    test('returns default value for null', () => {
        expect(getNumberValue(null, 99)).toBe(99);
    });
});

// sortNumbersArray

describe('sortNumbersArray', () => {
    test('sorts numbers in ascending order', () => {
        expect(sortNumbersArray([3, 1, 2])).toEqual([1, 2, 3]);
    });
});

// replaceAll

describe('replaceAll', () => {
    test('replaces all occurrences of a substring', () => {
        expect(replaceAll("hello world hello", "hello", "hi")).toBe("hi world hi");
    });
});

// removeAll

describe('removeAll', () => {
    test('removes all occurrences of a substring', () => {
        expect(removeAll("hello world", "hello")).toBe(" world");
    });
});
