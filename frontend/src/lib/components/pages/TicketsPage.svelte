<script lang="ts">
  import "../../../styles/tickets.sass";
  import { ConfirmationNumber } from "@material-symbols-svg/svelte";
  import AccountLayout from "../account/AccountLayout.svelte";
</script>

<div class="tickets-root">
  <AccountLayout title="My tickets">
    <div class="page-heading">
      <div>
        <p>Tickets for the events you've booked.</p>
      </div>
    </div>

    <div class="state">
      <ConfirmationNumber size={40} />
      <p>You don't have any tickets yet.</p>
      <a href="#/search" class="button-primary">Find events</a>
    </div>
  </AccountLayout>
</div>
