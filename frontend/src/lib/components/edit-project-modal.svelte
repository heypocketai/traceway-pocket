<script lang="ts">
    import * as Sheet from "$lib/components/ui/sheet";
    import * as AlertDialog from "$lib/components/ui/alert-dialog";
    import { Button } from "$lib/components/ui/button";
    import { Input } from "$lib/components/ui/input";
    import { Label } from "$lib/components/ui/label";
    import { projectsState, type Project, type Framework } from '$lib/state/projects.svelte';
    import { Check, Trash2 } from 'lucide-svelte';
    import FrameworkCombobox from './framework-combobox.svelte';
    import { toast } from 'svelte-sonner';
    import { goto } from '$app/navigation';

    interface Props {
        open: boolean;
        onOpenChange: (open: boolean) => void;
        project: Project | null;
    }

    let { open, onOpenChange, project }: Props = $props();

    let projectName = $state('');
    let selectedFramework = $state<Framework>('gin');
    let loading = $state(false);
    let error = $state('');
    let showDeleteConfirm = $state(false);
    let deleting = $state(false);

    $effect(() => {
        if (open && project) {
            projectName = project.name;
            selectedFramework = project.framework;
            error = '';
        }
    });

    async function handleSubmit(e: Event) {
        e.preventDefault();
        if (!projectName.trim()) {
            error = 'Project name is required';
            return;
        }
        if (!project) return;

        loading = true;
        error = '';

        try {
            await projectsState.updateProject(project.id, projectName.trim(), selectedFramework);
            toast.success('Successfully updated the project', { position: 'top-center' });
            onOpenChange(false);
        } catch (err: any) {
            error = err instanceof Error ? err.message : 'Failed to update project';
        } finally {
            loading = false;
        }
    }

    async function handleDelete() {
        if (!project) return;

        deleting = true;
        try {
            await projectsState.deleteProject(project.id);
            toast.success('Successfully deleted the project', { position: 'top-center' });
            showDeleteConfirm = false;
            onOpenChange(false);
            goto('/');
        } catch (err: any) {
            toast.error(err instanceof Error ? err.message : 'Failed to delete project', { position: 'top-center' });
        } finally {
            deleting = false;
        }
    }

    function handleClose() {
        error = '';
        showDeleteConfirm = false;
        onOpenChange(false);
    }
</script>

<Sheet.Root {open} onOpenChange={handleClose}>
    <Sheet.Content side="right" class="w-[400px] sm:w-[540px]">
        <Sheet.Header>
            <Sheet.Title>Edit Project</Sheet.Title>
            <Sheet.Description>
                Update your project name or framework.
            </Sheet.Description>
        </Sheet.Header>

        <form onsubmit={handleSubmit} class="px-6 py-6 space-y-5">
            <div class="space-y-2">
                <Label for="edit-project-name">Project Name</Label>
                <Input
                    id="edit-project-name"
                    type="text"
                    placeholder="My Application"
                    bind:value={projectName}
                    disabled={loading}
                />
                <p class="text-xs text-muted-foreground">
                    A unique name for your project (letters, numbers, spaces, hyphens)
                </p>
            </div>

            <div class="space-y-2">
                <Label for="edit-framework">Framework</Label>
                <FrameworkCombobox bind:value={selectedFramework} disabled={loading} />
                <p class="text-xs text-muted-foreground">
                    Select your framework for tailored integration code
                </p>
            </div>

            {#if error}
                <div class="rounded-md bg-destructive/10 border border-destructive/20 p-3">
                    <p class="text-sm text-destructive">{error}</p>
                </div>
            {/if}

            <div class="flex justify-end gap-2 pt-2">
                <Button type="button" variant="outline" onclick={handleClose} disabled={loading}>
                    Cancel
                </Button>
                <Button type="submit" disabled={loading}>
                    {#if loading}
                        Updating...
                    {:else}
                        <Check class="mr-2 h-4 w-4" />
                        Update Project
                    {/if}
                </Button>
            </div>
        </form>

        <div class="px-6 pb-6">
            <div class="border-t pt-6">
                <p class="text-sm text-muted-foreground mb-3">Danger zone</p>
                <Button
                    type="button"
                    variant="outline"
                    class="border-destructive text-destructive hover:bg-destructive hover:text-destructive-foreground"
                    onclick={() => showDeleteConfirm = true}
                >
                    <Trash2 class="mr-2 h-4 w-4" />
                    Delete Project
                </Button>
            </div>
        </div>
    </Sheet.Content>
</Sheet.Root>

<AlertDialog.Root bind:open={showDeleteConfirm}>
    <AlertDialog.Content interactOutsideBehavior="close">
        <AlertDialog.Header>
            <AlertDialog.Title>Delete Project</AlertDialog.Title>
            <AlertDialog.Description>
                Are you sure you want to delete this project? This action cannot be undone and all associated data will become inaccessible.
            </AlertDialog.Description>
        </AlertDialog.Header>
        <div class="rounded-md border px-4 py-3 mx-6">
            <p class="text-sm font-semibold">{project?.name}</p>
        </div>
        <AlertDialog.Footer>
            <Button variant="outline" onclick={() => showDeleteConfirm = false} disabled={deleting}>
                Cancel
            </Button>
            <Button variant="destructive" onclick={handleDelete} disabled={deleting}>
                {deleting ? 'Deleting...' : 'Delete Project'}
            </Button>
        </AlertDialog.Footer>
    </AlertDialog.Content>
</AlertDialog.Root>
