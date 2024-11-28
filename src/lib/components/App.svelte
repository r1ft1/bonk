<script lang="ts">
	import { Canvas } from "@threlte/core";
	import Scene from "./Scene.svelte";
	import Renderer from "./Renderer.svelte";
	import GameBrowser from "./GameBrowser.svelte";
	import { inGame } from "./stores.svelte";
	import GameInfo from "./GameInfo.svelte";
</script>

{#if !$inGame}
	<GameBrowser />
{:else}
	<GameInfo />
{/if}
<Canvas>
	<Renderer />
	<Scene />
</Canvas>
