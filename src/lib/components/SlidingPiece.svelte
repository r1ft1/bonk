<script lang="ts">
  import { Group, MathUtils } from "three";
  import { T, useTask } from "@threlte/core";
  import { useGltf, Outlines, Edges } from "@threlte/extras";
  import { isMobile, placementLanded, animConfig } from "./stores";

  let {
    tile: _tile,
    startPos: _startPos,
    endPos: _endPos,
    onDone,
  }: {
    tile: number;
    startPos: [number, number, number];
    endPos: [number, number, number];
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

  let elapsed = 0;
  let waitElapsed = 0;
  let started = false;

  function easeInOut(t: number): number {
    return t < 0.5 ? 2 * t * t : 1 - Math.pow(-2 * t + 2, 2) / 2;
  }

  useTask((delta) => {
    const cfg = $animConfig;
    // Wait for placement arc to land + configurable delay
    // Negative = skip ahead into animation (as if it started earlier)
    if (!started) {
      if (!$placementLanded) return;
      waitElapsed += delta;
      if (waitElapsed < cfg.slideDelay) return;
      started = true;
    }

    elapsed += delta;
    const t = easeInOut(Math.min(elapsed / cfg.slideDuration, 1));

    ref.position.x = MathUtils.lerp(startPos[0], endPos[0], t);
    ref.position.z = MathUtils.lerp(startPos[2], endPos[2], t);
    ref.position.y = MathUtils.lerp(startPos[1], endPos[1], t) + cfg.slideArcHeight * Math.sin(t * Math.PI);

    if (elapsed >= cfg.slideDuration) {
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
