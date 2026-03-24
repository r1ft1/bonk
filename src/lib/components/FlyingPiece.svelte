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
  let physicsAccum = 0;

  const PHYSICS_DT = 1 / 120; // Fixed 120Hz physics step

  function easeOut(t: number): number {
    return 1 - Math.pow(1 - t, 3);
  }

  function stepPhysics(dt: number, cfg: typeof $animConfig) {
    vy -= cfg.gravity * dt;
    ref.position.x += vx * dt;
    ref.position.z += vz * dt;
    ref.position.y += vy * dt;

    ref.rotation.x += dt * direction[1] * cfg.tumbleSpeed;
    ref.rotation.z -= dt * direction[0] * cfg.tumbleSpeed;

    if (ref.position.y <= cfg.groundY) {
      ref.position.y = cfg.groundY;
      if (Math.abs(vy) < cfg.bounceMinVelocity) {
        landed = true;
        landedTime = elapsed;
      } else {
        vy = Math.abs(vy) * cfg.bounceEnergyLoss;
        vx *= cfg.bounceFriction;
        vz *= cfg.bounceFriction;
      }
    }
  }

  let frameCount = 0;
  useTask((delta) => {
    const cfg = $animConfig;
    if (!started) {
      waitElapsed += delta;
      if (cfg.flyDelay < 0) {
        // Negative delay: start before placement lands
        if (waitElapsed < -cfg.flyDelay) return;
      } else {
        // Positive/zero delay: wait for landing + delay
        if (!$placementLanded) return;
        if (waitElapsed < cfg.flyDelay) return;
      }
      started = true;
      console.log('[FlyingPiece] started, first delta:', delta, 'waited:', waitElapsed.toFixed(3));
    }
    frameCount++;
    if (frameCount <= 5 || frameCount % 10 === 0) {
      console.log(`[FlyingPiece] f=${frameCount} dt=${delta.toFixed(4)} elapsed=${elapsed.toFixed(3)} y=${ref.position.y.toFixed(3)} vy=${vy.toFixed(2)} landed=${landed}`);
    }

    const edgeX = startPos[0] + direction[0] * cfg.bumpDistance;
    const edgeZ = startPos[2] + direction[1] * cfg.bumpDistance;

    elapsed += delta;

    // Phase 1: bump — slide to board edge with upward pop
    if (elapsed < cfg.bumpDuration) {
      const t = easeOut(elapsed / cfg.bumpDuration);
      ref.position.x = MathUtils.lerp(startPos[0], edgeX, t);
      ref.position.z = MathUtils.lerp(startPos[2], edgeZ, t);
      ref.position.y = startPos[1] + cfg.bumpArcHeight * Math.sin(t * Math.PI);
    }

    // Phase 2: physics-based freefall (fixed timestep)
    if (elapsed >= cfg.bumpDuration && !landed) {
      if (vy === 0 && vx === 0) {
        vy = cfg.bumpVelocityY;
        vx = direction[0] * cfg.bumpVelocityXZ;
        vz = direction[1] * cfg.bumpVelocityXZ;
      }

      physicsAccum += delta;
      while (physicsAccum >= PHYSICS_DT && !landed) {
        stepPhysics(PHYSICS_DT, cfg);
        physicsAccum -= PHYSICS_DT;
      }
    }

    // Phase 3: after landing, shrink away
    if (landed) {
      const t = Math.min((elapsed - landedTime) / cfg.shrinkDuration, 1);
      const s = 1 - t;
      ref.scale.set(s, s, s);
      if (t >= 1) onDone();
    }

    // Safety timeout
    if (elapsed >= 3) onDone();
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
