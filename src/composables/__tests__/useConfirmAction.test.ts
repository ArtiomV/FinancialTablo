import { describe, it, expect, jest } from '@jest/globals';
import { useConfirmAction } from '../useConfirmAction.ts';

// Mock vue ref — same pattern as useInfiniteScroll.test.ts
jest.unstable_mockModule('vue', () => ({
    ref: (val: any) => ({ value: val }),
    computed: (fn: any) => ({ get value() { return fn(); } })
}));

function makeConfirmDialog(shouldConfirm: boolean) {
    return {
        value: {
            open: jest.fn<() => Promise<void>>().mockImplementation(() =>
                shouldConfirm ? Promise.resolve() : Promise.reject(new Error('cancelled'))
            )
        }
    };
}

function makeSnackbar() {
    return {
        value: {
            showMessage: jest.fn(),
            showError: jest.fn()
        }
    };
}

describe('useConfirmAction', () => {
    it('should return isExecuting and confirmAndExecute', () => {
        const { isExecuting, confirmAndExecute } = useConfirmAction();
        expect(isExecuting).toBeDefined();
        expect(isExecuting.value).toBe(false);
        expect(typeof confirmAndExecute).toBe('function');
    });

    it('should do nothing when user cancels the dialog', async () => {
        const { confirmAndExecute, isExecuting } = useConfirmAction();
        const action = jest.fn<() => Promise<void>>().mockResolvedValue(undefined);
        const confirmDialog = makeConfirmDialog(false);
        const snackbar = makeSnackbar();

        await confirmAndExecute({
            confirmDialog: confirmDialog as any,
            snackbar: snackbar as any,
            title: 'Delete?',
            action,
            successMessage: 'Deleted'
        });

        expect(confirmDialog.value.open).toHaveBeenCalledWith('Delete?', undefined);
        expect(action).not.toHaveBeenCalled();
        expect(isExecuting.value).toBe(false);
    });

    it('should execute action and show success message on confirm', async () => {
        const { confirmAndExecute, isExecuting } = useConfirmAction();
        const action = jest.fn<() => Promise<void>>().mockResolvedValue(undefined);
        const confirmDialog = makeConfirmDialog(true);
        const snackbar = makeSnackbar();
        const onSuccess = jest.fn();

        await confirmAndExecute({
            confirmDialog: confirmDialog as any,
            snackbar: snackbar as any,
            title: 'Delete?',
            message: 'Are you sure?',
            action,
            successMessage: 'Deleted',
            onSuccess
        });

        expect(confirmDialog.value.open).toHaveBeenCalledWith('Delete?', 'Are you sure?');
        expect(action).toHaveBeenCalledTimes(1);
        expect(snackbar.value.showMessage).toHaveBeenCalledWith('Deleted');
        expect(onSuccess).toHaveBeenCalledTimes(1);
        expect(isExecuting.value).toBe(false);
    });

    it('should not show success message when successMessage is not provided', async () => {
        const { confirmAndExecute } = useConfirmAction();
        const action = jest.fn<() => Promise<void>>().mockResolvedValue(undefined);
        const confirmDialog = makeConfirmDialog(true);
        const snackbar = makeSnackbar();

        await confirmAndExecute({
            confirmDialog: confirmDialog as any,
            snackbar: snackbar as any,
            title: 'Confirm?',
            action
        });

        expect(action).toHaveBeenCalledTimes(1);
        expect(snackbar.value.showMessage).not.toHaveBeenCalled();
    });

    it('should show error message when action fails', async () => {
        const { confirmAndExecute, isExecuting } = useConfirmAction();
        const action = jest.fn<() => Promise<void>>().mockRejectedValue(new Error('Network error'));
        const confirmDialog = makeConfirmDialog(true);
        const snackbar = makeSnackbar();

        await confirmAndExecute({
            confirmDialog: confirmDialog as any,
            snackbar: snackbar as any,
            title: 'Delete?',
            action,
            errorMessage: 'Failed to delete'
        });

        expect(snackbar.value.showError).toHaveBeenCalledWith('Failed to delete');
        expect(isExecuting.value).toBe(false);
    });

    it('should use error.message when errorMessage is not provided', async () => {
        const { confirmAndExecute } = useConfirmAction();
        const action = jest.fn<() => Promise<void>>().mockRejectedValue(new Error('Something went wrong'));
        const confirmDialog = makeConfirmDialog(true);
        const snackbar = makeSnackbar();

        await confirmAndExecute({
            confirmDialog: confirmDialog as any,
            snackbar: snackbar as any,
            title: 'Delete?',
            action
        });

        expect(snackbar.value.showError).toHaveBeenCalledWith('Something went wrong');
    });

    it('should use fallback message when neither errorMessage nor error.message', async () => {
        const { confirmAndExecute } = useConfirmAction();
        const action = jest.fn<() => Promise<void>>().mockRejectedValue({});
        const confirmDialog = makeConfirmDialog(true);
        const snackbar = makeSnackbar();

        await confirmAndExecute({
            confirmDialog: confirmDialog as any,
            snackbar: snackbar as any,
            title: 'Delete?',
            action
        });

        expect(snackbar.value.showError).toHaveBeenCalledWith('An error occurred');
    });

    it('should not show error when error is marked as processed', async () => {
        const { confirmAndExecute } = useConfirmAction();
        const action = jest.fn<() => Promise<void>>().mockRejectedValue({ processed: true });
        const confirmDialog = makeConfirmDialog(true);
        const snackbar = makeSnackbar();

        await confirmAndExecute({
            confirmDialog: confirmDialog as any,
            snackbar: snackbar as any,
            title: 'Delete?',
            action,
            errorMessage: 'Failed'
        });

        expect(snackbar.value.showError).not.toHaveBeenCalled();
    });

    it('should set isExecuting to true during action execution', async () => {
        const { confirmAndExecute, isExecuting } = useConfirmAction();
        let executingDuringAction = false;

        const action = jest.fn<() => Promise<void>>().mockImplementation(async () => {
            executingDuringAction = isExecuting.value;
        });
        const confirmDialog = makeConfirmDialog(true);
        const snackbar = makeSnackbar();

        await confirmAndExecute({
            confirmDialog: confirmDialog as any,
            snackbar: snackbar as any,
            title: 'Test',
            action
        });

        expect(executingDuringAction).toBe(true);
        expect(isExecuting.value).toBe(false);
    });

    it('should handle undefined confirmDialog value gracefully', async () => {
        const { confirmAndExecute } = useConfirmAction();
        const action = jest.fn<() => Promise<void>>().mockResolvedValue(undefined);

        await confirmAndExecute({
            confirmDialog: { value: undefined } as any,
            snackbar: { value: undefined } as any,
            title: 'Test',
            action
        });

        // When confirmDialog.value is undefined, open() is never called,
        // await resolves, and action executes
        expect(action).toHaveBeenCalledTimes(1);
    });

    it('should handle undefined snackbar value on success', async () => {
        const { confirmAndExecute } = useConfirmAction();
        const action = jest.fn<() => Promise<void>>().mockResolvedValue(undefined);
        const confirmDialog = makeConfirmDialog(true);

        await confirmAndExecute({
            confirmDialog: confirmDialog as any,
            snackbar: { value: undefined } as any,
            title: 'Test',
            action,
            successMessage: 'Done'
        });

        expect(action).toHaveBeenCalledTimes(1);
    });

    it('should not call onSuccess when action fails', async () => {
        const { confirmAndExecute } = useConfirmAction();
        const action = jest.fn<() => Promise<void>>().mockRejectedValue(new Error('fail'));
        const confirmDialog = makeConfirmDialog(true);
        const snackbar = makeSnackbar();
        const onSuccess = jest.fn();

        await confirmAndExecute({
            confirmDialog: confirmDialog as any,
            snackbar: snackbar as any,
            title: 'Delete?',
            action,
            onSuccess
        });

        expect(onSuccess).not.toHaveBeenCalled();
    });

    it('should allow multiple sequential invocations', async () => {
        const { confirmAndExecute, isExecuting } = useConfirmAction();
        const action1 = jest.fn<() => Promise<void>>().mockResolvedValue(undefined);
        const action2 = jest.fn<() => Promise<void>>().mockResolvedValue(undefined);
        const confirmDialog = makeConfirmDialog(true);
        const snackbar = makeSnackbar();

        await confirmAndExecute({
            confirmDialog: confirmDialog as any,
            snackbar: snackbar as any,
            title: 'First?',
            action: action1
        });

        await confirmAndExecute({
            confirmDialog: confirmDialog as any,
            snackbar: snackbar as any,
            title: 'Second?',
            action: action2
        });

        expect(action1).toHaveBeenCalledTimes(1);
        expect(action2).toHaveBeenCalledTimes(1);
        expect(isExecuting.value).toBe(false);
    });
});
