import { describe, it, expect, beforeEach, jest } from '@jest/globals';
import { useInfiniteScroll } from '../useInfiniteScroll.ts';

// Mock IntersectionObserver
let observerCallback: IntersectionObserverCallback;
let observerOptions: IntersectionObserverInit | undefined;
const mockObserve = jest.fn();
const mockDisconnect = jest.fn();

(globalThis as any).IntersectionObserver = class MockIntersectionObserver {
    constructor(callback: IntersectionObserverCallback, options?: IntersectionObserverInit) {
        observerCallback = callback;
        observerOptions = options;
    }
    observe = mockObserve;
    disconnect = mockDisconnect;
    unobserve = jest.fn();
};

// Mock vue ref
jest.unstable_mockModule('vue', () => ({
    ref: (val: any) => ({ value: val })
}));

describe('useInfiniteScroll', () => {
    beforeEach(() => {
        mockObserve.mockClear();
        mockDisconnect.mockClear();
    });

    it('should return triggerRef, setupObserver, cleanup, and isLoading', () => {
        const { triggerRef, setupObserver, cleanup, isLoading } = useInfiniteScroll(
            () => {},
            () => true
        );
        expect(triggerRef).toBeDefined();
        expect(typeof setupObserver).toBe('function');
        expect(typeof cleanup).toBe('function');
        expect(isLoading).toBeDefined();
        expect(isLoading.value).toBe(false);
    });

    it('should not create observer when hasMore returns false', () => {
        const { triggerRef, setupObserver } = useInfiniteScroll(
            () => {},
            () => false
        );
        triggerRef.value = {} as HTMLElement;
        setupObserver();
        expect(mockObserve).not.toHaveBeenCalled();
    });

    it('should not create observer when triggerRef is null', () => {
        const { setupObserver } = useInfiniteScroll(
            () => {},
            () => true
        );
        setupObserver();
        expect(mockObserve).not.toHaveBeenCalled();
    });

    it('should create observer and observe element when hasMore and triggerRef set', () => {
        const { triggerRef, setupObserver } = useInfiniteScroll(
            () => {},
            () => true
        );
        const el = {} as HTMLElement;
        triggerRef.value = el;
        setupObserver();
        expect(mockObserve).toHaveBeenCalledTimes(1);
    });

    it('should use default threshold and rootMargin', () => {
        const { triggerRef, setupObserver } = useInfiniteScroll(
            () => {},
            () => true
        );
        triggerRef.value = {} as HTMLElement;
        setupObserver();
        expect(observerOptions?.threshold).toBe(0.1);
        expect(observerOptions?.rootMargin).toBe('0px');
    });

    it('should use custom threshold and rootMargin', () => {
        const { triggerRef, setupObserver } = useInfiniteScroll(
            () => {},
            () => true,
            { threshold: 0.5, rootMargin: '100px' }
        );
        triggerRef.value = {} as HTMLElement;
        setupObserver();
        expect(observerOptions?.threshold).toBe(0.5);
        expect(observerOptions?.rootMargin).toBe('100px');
    });

    it('should call loadMoreFn when entry is intersecting', () => {
        const loadMore = jest.fn();
        const { triggerRef, setupObserver, isLoading } = useInfiniteScroll(
            loadMore,
            () => true
        );
        triggerRef.value = {} as HTMLElement;
        setupObserver();

        // Simulate intersection
        observerCallback(
            [{ isIntersecting: true } as IntersectionObserverEntry],
            {} as IntersectionObserver
        );
        expect(loadMore).toHaveBeenCalledTimes(1);
    });

    it('should not call loadMoreFn when entry is not intersecting', () => {
        const loadMore = jest.fn();
        const { triggerRef, setupObserver } = useInfiniteScroll(
            loadMore,
            () => true
        );
        triggerRef.value = {} as HTMLElement;
        setupObserver();

        observerCallback(
            [{ isIntersecting: false } as IntersectionObserverEntry],
            {} as IntersectionObserver
        );
        expect(loadMore).not.toHaveBeenCalled();
    });

    it('should not call loadMoreFn when isLoading is true', () => {
        const loadMore = jest.fn();
        const { triggerRef, setupObserver, isLoading } = useInfiniteScroll(
            loadMore,
            () => true
        );
        triggerRef.value = {} as HTMLElement;
        setupObserver();

        isLoading.value = true;
        observerCallback(
            [{ isIntersecting: true } as IntersectionObserverEntry],
            {} as IntersectionObserver
        );
        expect(loadMore).not.toHaveBeenCalled();
    });

    it('should set isLoading to true during sync loadMoreFn and back to false', () => {
        let loadingDuringCall = false;
        const { triggerRef, setupObserver, isLoading } = useInfiniteScroll(
            () => { loadingDuringCall = isLoading.value; },
            () => true
        );
        triggerRef.value = {} as HTMLElement;
        setupObserver();

        observerCallback(
            [{ isIntersecting: true } as IntersectionObserverEntry],
            {} as IntersectionObserver
        );
        expect(loadingDuringCall).toBe(true);
        expect(isLoading.value).toBe(false);
    });

    it('should handle async loadMoreFn and set isLoading false after resolve', async () => {
        let resolveFn: () => void;
        const promise = new Promise<void>((resolve) => { resolveFn = resolve; });
        const { triggerRef, setupObserver, isLoading } = useInfiniteScroll(
            () => promise,
            () => true
        );
        triggerRef.value = {} as HTMLElement;
        setupObserver();

        observerCallback(
            [{ isIntersecting: true } as IntersectionObserverEntry],
            {} as IntersectionObserver
        );
        expect(isLoading.value).toBe(true);

        resolveFn!();
        await promise;
        // Wait for .finally() microtask
        await new Promise(resolve => setTimeout(resolve, 0));
        expect(isLoading.value).toBe(false);
    });

    it('should disconnect observer on cleanup', () => {
        const { triggerRef, setupObserver, cleanup } = useInfiniteScroll(
            () => {},
            () => true
        );
        triggerRef.value = {} as HTMLElement;
        setupObserver();
        cleanup();
        expect(mockDisconnect).toHaveBeenCalled();
    });

    it('should disconnect previous observer when setupObserver called again', () => {
        const { triggerRef, setupObserver } = useInfiniteScroll(
            () => {},
            () => true
        );
        triggerRef.value = {} as HTMLElement;
        setupObserver();
        setupObserver();
        expect(mockDisconnect).toHaveBeenCalledTimes(1);
    });

    it('should handle empty entries array', () => {
        const loadMore = jest.fn();
        const { triggerRef, setupObserver } = useInfiniteScroll(
            loadMore,
            () => true
        );
        triggerRef.value = {} as HTMLElement;
        setupObserver();

        observerCallback([], {} as IntersectionObserver);
        expect(loadMore).not.toHaveBeenCalled();
    });
});
