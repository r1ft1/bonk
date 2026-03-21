<script lang="ts">
  import { Group, MathUtils } from "three";
  import { T, useTask } from "@threlte/core";
  import { useGltf, Outlines, Edges } from "@threlte/extras";
  import { isMobile } from "./stores";

  let {
    positions: _positions,
    tile: _tile,
    onDone,
  }: {
    positions: [number, number, number][];
    tile: number;
    onDone: () => void;
  } = $props();

  // Snapshot props — these are initial values for a one-shot animation
  const positions = _positions.map(p => [...p] as [number, number, number]);
  const tile = _tile;

  const color = tile === 1 ? "orange" : "lightblue";
  const isP1 = tile === 1;
  const cornerX = isP1 ? -4 : 4;
  const cornerZ = isP1 ? -4 : 4;

  const kittenGltf = useGltf("/kitten.glb");
  const catGltf = useGltf("/cat.glb");

  const ref0 = new Group();
  const ref1 = new Group();
  const ref2 = new Group();
  const catRef = new Group();

  ref0.position.set(positions[0][0], positions[0][1], positions[0][2]);
  ref1.position.set(positions[1][0], positions[1][1], positions[1][2]);
  ref2.position.set(positions[2][0], positions[2][1], positions[2][2]);
  catRef.position.set(positions[1][0], positions[1][1], positions[1][2]);
  catRef.scale.set(0, 0, 0);

  const middle = positions[1];

  // Timeline-driven animation using useTask (not affected by Reduced Motion)
  let elapsed = 0;

  // Snapshot start positions for lerping
  const start0 = { x: positions[0][0], z: positions[0][2] };
  const start2 = { x: positions[2][0], z: positions[2][2] };

  function easeInOut(t: number): number {
    return t < 0.5 ? 2 * t * t : 1 - Math.pow(-2 * t + 2, 2) / 2;
  }

  function easeOut(t: number): number {
    return 1 - Math.pow(1 - t, 3);
  }

  useTask((delta) => {
    elapsed += delta;

    // Phase 1 (0–0.35s): outer kittens slide toward middle
    if (elapsed < 0.35) {
      const t = easeInOut(Math.min(elapsed / 0.35, 1));
      ref0.position.x = MathUtils.lerp(start0.x, middle[0], t);
      ref0.position.z = MathUtils.lerp(start0.z, middle[2], t);
      ref2.position.x = MathUtils.lerp(start2.x, middle[0], t);
      ref2.position.z = MathUtils.lerp(start2.z, middle[2], t);
    } else if (elapsed < 0.45) {
      // Phase 1 done — snap to middle
      ref0.position.x = middle[0];
      ref0.position.z = middle[2];
      ref2.position.x = middle[0];
      ref2.position.z = middle[2];
    }

    // Phase 2 (0.35–0.6s): kittens shrink, cat pops up
    if (elapsed >= 0.35 && elapsed < 0.6) {
      const shrinkT = Math.min((elapsed - 0.35) / 0.1, 1);
      const s = 1 - shrinkT;
      ref0.scale.set(s, s, s);
      ref1.scale.set(s, s, s);
      ref2.scale.set(s, s, s);

      const popT = easeOut(Math.min((elapsed - 0.35) / 0.25, 1));
      // Scale: 0 → 1.3 → 1 (overshoot then settle)
      const catScale = popT < 0.6 ? (popT / 0.6) * 1.3 : 1.3 - 0.3 * ((popT - 0.6) / 0.4);
      catRef.scale.set(catScale, catScale, catScale);
    } else if (elapsed >= 0.6) {
      // Ensure final states
      ref0.scale.set(0, 0, 0);
      ref1.scale.set(0, 0, 0);
      ref2.scale.set(0, 0, 0);
      catRef.scale.set(1, 1, 1);
    }

    // Phase 3 (0.6–0.75s): cat hops up
    if (elapsed >= 0.6 && elapsed < 0.75) {
      const hopT = easeOut((elapsed - 0.6) / 0.15);
      catRef.position.y = middle[1] + 0.8 * hopT;
    }

    // Phase 3b (0.75–1.45s): cat flies to corner
    if (elapsed >= 0.75 && elapsed < 1.45) {
      const flyT = easeInOut((elapsed - 0.75) / 0.7);
      const hopY = middle[1] + 0.8; // starting y after hop
      catRef.position.x = MathUtils.lerp(middle[0], cornerX, flyT);
      catRef.position.y = MathUtils.lerp(hopY, 0.52, flyT);
      catRef.position.z = MathUtils.lerp(middle[2], cornerZ, flyT);
    }

    // Done
    if (elapsed >= 1.45) {
      onDone();
    }
  });
</script>

{#await Promise.all([kittenGltf, catGltf]) then [kg, cg]}
  <T is={ref0} dispose={false}>
    <T.Mesh geometry={kg.nodes.Kitten.geometry} scale={[0.5, 0.5, 0.5]} rotation.y={$isMobile ? -Math.PI / 2 : 0}>
      <T.MeshStandardMaterial {color} />
      <Outlines color="black" />
    </T.Mesh>
  </T>
  <T is={ref1} dispose={false}>
    <T.Mesh geometry={kg.nodes.Kitten.geometry} scale={[0.5, 0.5, 0.5]} rotation.y={$isMobile ? -Math.PI / 2 : 0}>
      <T.MeshStandardMaterial {color} />
      <Outlines color="black" />
    </T.Mesh>
  </T>
  <T is={ref2} dispose={false}>
    <T.Mesh geometry={kg.nodes.Kitten.geometry} scale={[0.5, 0.5, 0.5]} rotation.y={$isMobile ? -Math.PI / 2 : 0}>
      <T.MeshStandardMaterial {color} />
      <Outlines color="black" />
    </T.Mesh>
  </T>
  <T is={catRef} dispose={false}>
    <T.Mesh geometry={cg.nodes.Cube.geometry} scale={[0.5, 0.5, 0.5]} rotation.y={$isMobile ? -Math.PI / 2 : 0}>
      <T.MeshStandardMaterial {color} />
      <Outlines color="black" />
      <Edges color="black" />
    </T.Mesh>
  </T>
{/await}
