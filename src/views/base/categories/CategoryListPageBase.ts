import { ref } from 'vue';

export function useCategoryListPageBase() {
    const loading = ref<boolean>(true);

    return {
        // states
        loading
    };
}
