<script lang="ts">
  import { Group, MathUtils } from "three";
  import { T, useTask } from "@threlte/core";
  import { useGltf, Outlines, Edges } from "@threlte/extras";
  import { isMobile } from "./stores";

  let {
    tile: _tile,
    startPos: _startPos,
    endPos: _endPos,
    delay = 0.3,
    onDone,
  }: {
    tile: number;
    startPos: [number, number, number];
    endPos: [number, number, number];
    delay?: number;
    onDone: () => void;
  } = $props();

  // Snapshot props — these are initial values for a one-shot animation
  const tile = _tile;
  const startPos = [..._startPos] as [number, number, number];
  const endPos = [..._endPos] as [number, number, number];

  const ref = new Group();
  ref.position.set(startPos[0], startPos[1], startPos[2]);

  const isCat = tile === 2 || tile === 9;
  const color = tile === 1 || tile === 2 ? "orange" : "lightblue";
  const gltf = useGltf(isCat ? "/cat.glb" : "/kitten.glb");

  const duration = 0.25;
  let elapsed = 0;

  function easeInOut(t: number): number {
    return t < 0.5 ? 2 * t * t : 1 - Math.pow(-2 * t + 2, 2) / 2;
  }

  useTask((delta) => {
    elapsed += delta;

    // Wait for delay before starting the slide
    if (elapsed < delay) return;

    const moveElapsed = elapsed - delay;
    const t = easeInOut(Math.min(moveElapsed / duration, 1));

    ref.position.x = MathUtils.lerp(startPos[0], endPos[0], t);
    ref.position.z = MathUtils.lerp(startPos[2], endPos[2], t);
    // Small arc upward during slide
    ref.position.y = MathUtils.lerp(startPos[1], endPos[1], t) + 0.2 * Math.sin(t * Math.PI);

    if (moveElapsed >= duration) {
      onDone();
    }
  });
</script>

{#await gltf then gltf}
  <T is={ref} dispose={false}>
    {#if isCat}
      <T.Mesh geometry={gltf.nodes.Cube.geometry} scale={[0.5, 0.5, 0.5]} rotation.y={$isMobile ? -Math.PI / 2 : 0}>
        <T.MeshStandardMaterial {color} />
        <Outlines color="black" />
        <Edges color="black" />
      </T.Mesh>
    {:else}
      <T.Mesh geometry={gltf.nodes.Kitten.geometry} scale={[0.5, 0.5, 0.5]} rotation.y={$isMobile ? -Math.PI / 2 : 0}>
        <T.MeshStandardMaterial {color} />
        <Outlines color="black" />
      </T.Mesh>
    {/if}
  </T>
{/await}
