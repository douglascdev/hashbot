<script>
  export let className = ''; // Optional class name for additional styling
</script>

<div class={`container mx-auto max-w-1x1 px-4 ${className} border border-base-300 rounded-lg p-1`}>
  <slot></slot> <!-- This is where child components will be rendered -->
</div>

<style>
  /* You can add additional styles here if needed */
</style>
