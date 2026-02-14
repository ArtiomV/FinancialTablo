/**
 * Composable for confirm-then-act patterns.
 *
 * Encapsulates the common pattern of showing a confirmation dialog
 * before performing a destructive action (delete, confirm planned
 * transaction, etc.).
 *
 * Usage:
 * ```ts
 * const { confirmAndExecute, isExecuting } = useConfirmAction();
 *
 * async function deleteItem(item: Item) {
 *   await confirmAndExecute({
 *     confirmDialog,
 *     snackbar,
 *     title: 'Delete this item?',
 *     action: () => store.deleteItem(item),
 *     successMessage: 'Item deleted',
 *     errorMessage: 'Failed to delete item',
 *     onSuccess: () => reload()
 *   });
 * }
 * ```
 */

import { ref, type Ref } from 'vue';

interface ConfirmDialogRef {
    open: (title: string, message?: string) => Promise<void>;
}

interface SnackBarRef {
    showMessage: (message: string) => void;
    showError: (error: string | { message?: string }) => void;
}

export interface ConfirmAndExecuteOptions {
    /** Reference to the ConfirmDialog component */
    confirmDialog: Ref<ConfirmDialogRef | undefined>;
    /** Reference to the SnackBar component */
    snackbar: Ref<SnackBarRef | undefined>;
    /** Confirmation dialog title */
    title: string;
    /** Optional confirmation dialog message body */
    message?: string;
    /** The async action to perform after confirmation */
    action: () => Promise<unknown>;
    /** Message shown on success */
    successMessage?: string;
    /** Message shown on error */
    errorMessage?: string;
    /** Callback after successful action */
    onSuccess?: () => void;
}

export interface UseConfirmActionReturn {
    /** Whether an action is currently executing */
    isExecuting: Ref<boolean>;
    /** Show confirmation dialog, then execute action */
    confirmAndExecute: (options: ConfirmAndExecuteOptions) => Promise<void>;
}

export function useConfirmAction(): UseConfirmActionReturn {
    const isExecuting = ref<boolean>(false);

    async function confirmAndExecute(options: ConfirmAndExecuteOptions): Promise<void> {
        try {
            await options.confirmDialog.value?.open(options.title, options.message);
        } catch {
            // User cancelled the dialog
            return;
        }

        isExecuting.value = true;

        try {
            await options.action();
            isExecuting.value = false;

            if (options.successMessage) {
                options.snackbar.value?.showMessage(options.successMessage);
            }

            options.onSuccess?.();
        } catch (error: unknown) {
            isExecuting.value = false;

            const errorObj = error as { processed?: boolean; message?: string };
            if (!errorObj.processed) {
                options.snackbar.value?.showError(
                    options.errorMessage || errorObj.message || 'An error occurred'
                );
            }
        }
    }

    return {
        isExecuting,
        confirmAndExecute
    };
}
