<script lang='ts'>
    import Kitten from "./Kitten.svelte";
    import Cat from "./Cat.svelte";

    export let piece;
    export let placed;
    export let position;
    export let finalPosition=[0,0,0];
    export let booped;

    let color:string = "orange";

    let kitten=true
    if (piece == 1 || piece == 8) {
        kitten = true;
    } else if (piece == 2 || piece == 9) {
        kitten = false;
    }

    if (piece == 1 || piece == 2) {
        color = "orange";
    } else if (piece == 8 || piece == 9) {
        color = "lightblue";
    }
</script>

{#if kitten}
    <Kitten color={color} placed={placed} position={position} finalPosition={finalPosition} booped={booped}/>
{:else}
    <Cat color={color} placed={placed} position={position} finalPosition={finalPosition} booped={booped}/>
{/if}