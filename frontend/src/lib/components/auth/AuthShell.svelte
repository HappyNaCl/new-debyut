<script lang="ts">
  import type { Snippet } from "svelte";
  import AuthIllustration from "../ui/AuthIllustration.svelte";

  interface Props {
    children: Snippet;
  }

  let { children }: Props = $props();
</script>

<header class="auth-header">
  <a href="#/" class="wordmark">debyut</a>
</header>

<div class="auth-body">
  <section class="auth-hero">
    <AuthIllustration />
    <h2>Never miss your favorite events again</h2>
    <p>Join debyut and enjoy the easiest way to buy tickets and manage your events.</p>
  </section>

  <section class="auth-panel">
    {@render children()}
  </section>
</div>
