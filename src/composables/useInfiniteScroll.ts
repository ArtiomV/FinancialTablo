/**
 * Composable for infinite scroll pagination.
 *
 * Provides a reusable IntersectionObserver-based infinite scroll mechanism.
 * Extracted from transaction ListPage to avoid duplicating this pattern
 * across multiple list views (transactions, assets, obligations, etc.).
 *
 * Usage:
 * ```ts
 * const { triggerRef, setupObserver, cleanup } = useInfiniteScroll(() => loadMore());
 *
 * // In template: <div ref="triggerRef" v-if="hasMore" />
 * // After data loads: nextTick(() => setupObserver());
 * // In onBeforeUnmount: cleanup();
 * ```
 */

import { ref, type Ref } from 'vue';

export interface UseInfiniteScrollOptions {
    /** IntersectionObserver threshold (0-1). Default: 0.1 */
    threshold?: number;
    /** Additional root margin. Default: '0px' */
    rootMargin?: string;
}

export interface UseInfiniteScrollReturn {
    /** Template ref for the trigger element */
    triggerRef: Ref<HTMLElement | null>;
    /** Set up the observer (call after data loads in nextTick) */
    setupObserver: () => void;
    /** Clean up the observer (call in onBeforeUnmount) */
    cleanup: () => void;
    /** Whether loading is in progress */
    isLoading: Ref<boolean>;
}

export function useInfiniteScroll(
    loadMoreFn: () => void | Promise<void>,
    hasMore: () => boolean,
    options?: UseInfiniteScrollOptions
): UseInfiniteScrollReturn {
    const triggerRef = ref<HTMLElement | null>(null);
    const isLoading = ref<boolean>(false);
    let observer: IntersectionObserver | null = null;

    function setupObserver(): void {
        // Clean up previous observer
        if (observer) {
            observer.disconnect();
            observer = null;
        }

        if (!hasMore() || !triggerRef.value) {
            return;
        }

        observer = new IntersectionObserver(
            (entries) => {
                if (entries.length > 0 && entries[0]!.isIntersecting && !isLoading.value) {
                    isLoading.value = true;
                    const result = loadMoreFn();
                    if (result instanceof Promise) {
                        result.finally(() => {
                            isLoading.value = false;
                        });
                    } else {
                        isLoading.value = false;
                    }
                }
            },
            {
                threshold: options?.threshold ?? 0.1,
                rootMargin: options?.rootMargin ?? '0px'
            }
        );

        observer.observe(triggerRef.value);
    }

    function cleanup(): void {
        if (observer) {
            observer.disconnect();
            observer = null;
        }
    }

    return {
        triggerRef,
        setupObserver,
        cleanup,
        isLoading
    };
}
