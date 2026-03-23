<script lang="ts">
  import { Group, MathUtils } from "three";
  import { T, useTask } from "@threlte/core";
  import { useGltf, Outlines, Edges } from "@threlte/extras";
  import { isMobile, placementLanded, animConfig } from "./stores";

  let {
    tile: _tile,
    startPos: _startPos,
    direction: _direction,
    onDone,
  }: {
    tile: number;
    startPos: [number, number, number];
    direction: [number, number];
    onDone: () => void;
  } = $props();

  // Snapshot props — these are initial values for a one-shot animation
  const tile = _tile;
  const startPos = [..._startPos] as [number, number, number];
  const direction = [..._direction] as [number, number];

  const ref = new Group();
  ref.position.set(startPos[0], startPos[1], startPos[2]);

  const isCat = tile === 2 || tile === 9;
  const color = tile === 1 || tile === 2 ? "orange" : "lightblue";
  const gltf = useGltf(isCat ? "/cat.glb" : "/kitten.glb");

  let elapsed = 0;
  let waitElapsed = 0;
  let vy = 0;
  let vx = 0;
  let vz = 0;
  let landed = false;
  let landedTime = 0;
  let started = false;

  function easeOut(t: number): number {
    return 1 - Math.pow(1 - t, 3);
  }

  useTask((delta) => {
    const cfg = $animConfig;
    // Wait for placement arc to land + configurable delay
    // Negative = skip ahead into animation (as if it started earlier)
    if (!started) {
      if (!$placementLanded) return;
      waitElapsed += delta;
      if (waitElapsed < cfg.flyDelay) return;
      started = true;
    }

    const edgeX = startPos[0] + direction[0] * cfg.bumpDistance;
    const edgeZ = startPos[2] + direction[1] * cfg.bumpDistance;

    elapsed += delta;
    const t_elapsed = elapsed;

    // Phase 1: bump — slide to board edge with upward pop
    if (t_elapsed < cfg.bumpDuration) {
      const t = easeOut(t_elapsed / cfg.bumpDuration);
      ref.position.x = MathUtils.lerp(startPos[0], edgeX, t);
      ref.position.z = MathUtils.lerp(startPos[2], edgeZ, t);
      ref.position.y = startPos[1] + cfg.bumpArcHeight * Math.sin(t * Math.PI);
    }

    // Phase 2: physics-based freefall
    if (t_elapsed >= cfg.bumpDuration && !landed) {
      if (vy === 0 && vx === 0) {
        vy = cfg.bumpVelocityY;
        vx = direction[0] * cfg.bumpVelocityXZ;
        vz = direction[1] * cfg.bumpVelocityXZ;
      }

      vy -= cfg.gravity * delta;
      ref.position.x += vx * delta;
      ref.position.z += vz * delta;
      ref.position.y += vy * delta;

      // Tumble
      ref.rotation.x += delta * direction[1] * cfg.tumbleSpeed;
      ref.rotation.z -= delta * direction[0] * cfg.tumbleSpeed;

      // Hit the ground — bounce
      if (ref.position.y <= cfg.groundY) {
        ref.position.y = cfg.groundY;
        if (Math.abs(vy) < cfg.bounceMinVelocity) {
          landed = true;
          landedTime = t_elapsed;
        } else {
          vy = Math.abs(vy) * cfg.bounceEnergyLoss;
          vx *= cfg.bounceFriction;
          vz *= cfg.bounceFriction;
        }
      }
    }

    // Phase 3: after landing, shrink away
    if (landed) {
      const t = Math.min((t_elapsed - landedTime) / cfg.shrinkDuration, 1);
      const s = 1 - t;
      ref.scale.set(s, s, s);
      if (t >= 1) onDone();
    }

    // Safety timeout
    if (t_elapsed >= 3) onDone();
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
