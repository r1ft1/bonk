<script lang="ts">
  import { Group, MathUtils } from "three";
  import { T, useTask } from "@threlte/core";
  import { useGltf, Outlines, Edges } from "@threlte/extras";

  let {
    tile,
    startPos,
    direction,
    onDone,
  }: {
    tile: number;
    startPos: [number, number, number];
    direction: [number, number];
    onDone: () => void;
  } = $props();

  const ref = new Group();
  ref.position.set(startPos[0], startPos[1], startPos[2]);

  const isCat = tile === 2 || tile === 9;
  const color = tile === 1 || tile === 2 ? "orange" : "lightblue";
  const gltf = useGltf(isCat ? "/cat.glb" : "/kitten.glb");

  // Fly 4 units in the boop direction (world space: dx maps to x, dy maps to z)
  const endX = startPos[0] + direction[0] * 4;
  const endZ = startPos[2] + direction[1] * 4;

  let elapsed = 0;

  function easeIn(t: number): number {
    return t * t;
  }

  useTask((delta) => {
    elapsed += delta;

    // Phase 1 (0–0.1s): slight hop up
    if (elapsed < 0.1) {
      const t = elapsed / 0.1;
      ref.position.y = startPos[1] + 0.4 * t;
    }

    // Phase 2 (0.1–0.6s): fly off in boop direction, accelerating
    if (elapsed >= 0.1 && elapsed < 0.6) {
      const t = easeIn((elapsed - 0.1) / 0.5);
      ref.position.x = MathUtils.lerp(startPos[0], endX, t);
      ref.position.y = MathUtils.lerp(startPos[1] + 0.4, startPos[1], t);
      ref.position.z = MathUtils.lerp(startPos[2], endZ, t);
      // Shrink as it flies away
      const s = 1 - t * 0.5;
      ref.scale.set(s, s, s);
    }

    if (elapsed >= 0.6) {
      onDone();
    }
  });
</script>

{#await gltf then gltf}
  <T is={ref} dispose={false}>
    {#if isCat}
      <T.Mesh geometry={gltf.nodes.Cube.geometry} scale={[0.5, 0.5, 0.5]}>
        <T.MeshStandardMaterial {color} />
        <Outlines color="black" />
        <Edges color="black" />
      </T.Mesh>
    {:else}
      <T.Mesh geometry={gltf.nodes.Kitten.geometry} scale={[0.5, 0.5, 0.5]}>
        <T.MeshStandardMaterial {color} />
        <Outlines color="black" />
      </T.Mesh>
    {/if}
  </T>
{/await}
