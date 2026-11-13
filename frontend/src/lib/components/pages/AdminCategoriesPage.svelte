<script lang="ts">
  import "../../../styles/admin.sass";
  import { Add, Delete, Edit } from "@material-symbols-svg/svelte";
  import AdminLayout from "../admin/AdminLayout.svelte";
  import CategoryForm from "../admin/CategoryForm.svelte";
  import Modal from "../ui/Modal.svelte";
  import Spinner from "../ui/Spinner.svelte";
  import { extractErrorMessage } from "../../api/client";
  import {
    categoriesQueryKey,
    createCategoriesQuery,
    createCategory,
    deleteCategory,
    updateCategory,
  } from "../../api/categories";
  import { getCategoryIcon } from "../../icons/categoryIcons";
  import { queryClient } from "../../queryClient";
  import type { CategoryInput, EventCategory } from "../../types/category";

  type Editor = { mode: "create" } | { mode: "edit"; category: EventCategory } | null;

  const categories = createCategoriesQuery();

  let editor = $state<Editor>(null);
  let pendingDelete = $state<EventCategory | null>(null);
  let deleting = $state(false);
  let errorMessage = $state("");
  let successMessage = $state("");

  function refresh() {
    return queryClient.invalidateQueries({ queryKey: categoriesQueryKey });
  }

  function openEditor(next: Editor) {
    editor = next;
    errorMessage = "";
    successMessage = "";
  }

  async function handleSubmit(input: CategoryInput) {
    if (editor?.mode === "edit") {
      const updated = await updateCategory(editor.category.id, input);
      successMessage = `Saved ${updated.name}.`;
    } else {
      const created = await createCategory(input);
      successMessage = `Created ${created.name}.`;
    }
    editor = null;
    await refresh();
  }

  function askDelete(category: EventCategory) {
    pendingDelete = category;
    errorMessage = "";
    successMessage = "";
  }

  function closeDelete() {
    if (!deleting) pendingDelete = null;
  }

  async function handleDelete() {
    const category = pendingDelete;
    if (!category) return;
    errorMessage = "";
    successMessage = "";
    deleting = true;
    try {
      await deleteCategory(category.id);
      successMessage = `Deleted ${category.name}.`;
      if (editor?.mode === "edit" && editor.category.id === category.id) editor = null;
      await refresh();
    } catch (error) {
      errorMessage = extractErrorMessage(error);
    } finally {
      deleting = false;
      pendingDelete = null;
    }
  }
</script>

<div class="admin-root">
  <AdminLayout title="Event categories">
    <div class="page-heading">
      <div>
        <p>Categories appear in the header menu and on event pages.</p>
      </div>
      <button type="button" class="button-primary" onclick={() => openEditor({ mode: "create" })}>
        <Add size={20} /> New category
      </button>
    </div>

    {#if successMessage}
      <div class="alert alert-success" role="status">{successMessage}</div>
    {/if}
    {#if errorMessage}
      <div class="alert alert-error" role="alert">{errorMessage}</div>
    {/if}

    <Modal open={editor !== null} onClose={() => (editor = null)}>
      {#if editor}
        {#key editor}
          <CategoryForm
            category={editor.mode === "edit" ? editor.category : undefined}
            onSubmit={handleSubmit}
            onCancel={() => (editor = null)}
          />
        {/key}
      {/if}
    </Modal>

    <Modal open={pendingDelete !== null} title="Delete category" size="small" dismissible={!deleting} onClose={closeDelete}>
      {#if pendingDelete}
        <p class="confirm-message">
          Delete <strong>{pendingDelete.name}</strong>? This can't be undone.
        </p>
        <div class="confirm-actions">
          <button type="button" class="button-outline" onclick={closeDelete} disabled={deleting}>Cancel</button>
          <button type="button" class="button-danger" onclick={handleDelete} disabled={deleting}>
            {deleting ? "Deleting…" : "Delete"}
          </button>
        </div>
      {/if}
    </Modal>

    {#if categories.isPending}
      <div class="state"><Spinner size={32} /></div>
    {:else if categories.isError}
      <div class="state">
        <p>{extractErrorMessage(categories.error)}</p>
        <button type="button" class="button-outline" onclick={() => categories.refetch()}>Try again</button>
      </div>
    {:else if categories.data.length === 0}
      <div class="state">
        <p>No categories yet. Create the first one.</p>
      </div>
    {:else}
      <ul class="item-list">
        {#each categories.data as category (category.id)}
          {@const Icon = getCategoryIcon(category.icon)}
          <li class="item-row">
            <span class="item-icon"><Icon size={22} /></span>
            <div class="item-text">
              <span class="item-title">{category.name}</span>
              <span class="item-detail">{category.slug}</span>
            </div>

            <div class="row-actions">
              <button
                type="button"
                class="icon-button"
                aria-label={`Edit ${category.name}`}
                title="Edit"
                onclick={() => openEditor({ mode: "edit", category })}
              >
                <Edit size={20} />
              </button>
              <button
                type="button"
                class="icon-button danger"
                aria-label={`Delete ${category.name}`}
                title="Delete"
                onclick={() => askDelete(category)}
              >
                <Delete size={20} />
              </button>
            </div>
          </li>
        {/each}
      </ul>
    {/if}
  </AdminLayout>
</div>

<style lang="sass">
  .confirm-message
    margin: 0
    font-size: 14px
    line-height: 1.5
    color: var(--color-text)
    overflow-wrap: anywhere

  .confirm-actions
    display: flex
    justify-content: flex-end
    gap: 8px
</style>
