/**
 * Transaction filter parameter utilities.
 *
 * These helpers build URL query strings and export request objects
 * from the transaction list filter state, keeping this logic
 * separate from the Pinia store for testability and reuse.
 */

import { DateRange } from '@/core/datetime.ts';
import type { TransactionListFilter } from '@/stores/transaction.ts';
import type { ExportTransactionDataRequest } from '@/models/data_management.ts';

/** Default page type (List = 0). Avoids importing TransactionListPageBase which pulls in Vue. */
const DEFAULT_PAGE_TYPE = 0;

/**
 * Build a URL query string from the current transaction list filter.
 */
export function buildTransactionListPageParams(filter: TransactionListFilter, pageType: number): string {
    const params: string[] = [];

    if (pageType !== DEFAULT_PAGE_TYPE) {
        params.push(`pageType=${pageType}`);
    }

    if (filter.dateType && filter.dateType !== DateRange.All.type) {
        params.push(`dateType=${filter.dateType}`);
    }

    if (filter.minTime) {
        params.push(`minTime=${filter.minTime}`);
    }

    if (filter.maxTime) {
        params.push(`maxTime=${filter.maxTime}`);
    }

    if (filter.type) {
        params.push(`type=${filter.type}`);
    }

    if (filter.categoryIds) {
        params.push(`categoryIds=${encodeURIComponent(filter.categoryIds)}`);
    }

    if (filter.accountIds) {
        params.push(`accountIds=${encodeURIComponent(filter.accountIds)}`);
    }

    if (filter.counterpartyId) {
        params.push(`counterpartyId=${encodeURIComponent(filter.counterpartyId)}`);
    }

    if (filter.tagFilter) {
        params.push(`tagFilter=${encodeURIComponent(filter.tagFilter)}`);
    }

    if (filter.amountFilter) {
        params.push(`amountFilter=${encodeURIComponent(filter.amountFilter)}`);
    }

    if (filter.keyword) {
        params.push(`keyword=${encodeURIComponent(filter.keyword)}`);
    }

    return params.join('&');
}

/**
 * Build an export data request from the current transaction list filter.
 */
export function buildExportRequestFromFilter(filter: TransactionListFilter): ExportTransactionDataRequest {
    return {
        maxTime: filter.maxTime,
        minTime: filter.minTime,
        type: filter.type,
        categoryIds: filter.categoryIds,
        accountIds: filter.accountIds,
        counterpartyId: filter.counterpartyId,
        tagFilter: filter.tagFilter,
        amountFilter: filter.amountFilter,
        keyword: filter.keyword
    };
}
