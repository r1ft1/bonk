<script lang="ts">
  import { Group, MathUtils } from "three";
  import { T, useTask } from "@threlte/core";
  import { useGltf, Outlines, Edges } from "@threlte/extras";
  import { isMobile } from "./stores";

  let {
    tile: _tile,
    startPos: _startPos,
    direction: _direction,
    delay = 0,
    onDone,
  }: {
    tile: number;
    startPos: [number, number, number];
    direction: [number, number];
    delay?: number;
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

  // Bumped off the edge of the board like falling off a bed
  const edgeX = startPos[0] + direction[0] * 1.5;
  const edgeZ = startPos[2] + direction[1] * 1.5;
  const groundY = -2.7;

  // Physics constants
  const gravity = 18; // units/s^2
  const bumpVelocityY = 3; // initial upward pop
  const bumpVelocityXZ = 4; // horizontal bump speed

  let elapsed = 0;
  let vy = 0; // vertical velocity
  let vx = 0; // horizontal velocities
  let vz = 0;
  let landed = false;
  let landedTime = 0;

  function easeOut(t: number): number {
    return 1 - Math.pow(1 - t, 3);
  }

  useTask((delta) => {
    elapsed += delta;

    // Wait at start position during delay
    if (elapsed < delay) return;
    const t_elapsed = elapsed - delay;

    // Phase 1 (0–0.15s): bump — slide to board edge with upward pop
    if (t_elapsed < 0.15) {
      const t = easeOut(t_elapsed / 0.15);
      ref.position.x = MathUtils.lerp(startPos[0], edgeX, t);
      ref.position.z = MathUtils.lerp(startPos[2], edgeZ, t);
      ref.position.y = startPos[1] + 0.3 * Math.sin(t * Math.PI);
    }

    // Phase 2 (0.15s+): physics-based freefall
    if (t_elapsed >= 0.15 && !landed) {
      // Initialize velocity on first frame of freefall
      if (vy === 0 && vx === 0) {
        vy = bumpVelocityY;
        vx = direction[0] * bumpVelocityXZ;
        vz = direction[1] * bumpVelocityXZ;
      }

      // Apply gravity
      vy -= gravity * delta;

      // Update position
      ref.position.x += vx * delta;
      ref.position.z += vz * delta;
      ref.position.y += vy * delta;

      // Tumble
      ref.rotation.x += delta * direction[1] * 6;
      ref.rotation.z -= delta * direction[0] * 6;

      // Hit the ground — bounce
      if (ref.position.y <= groundY) {
        ref.position.y = groundY;
        if (Math.abs(vy) < 1) {
          // Too slow to bounce, just land
          landed = true;
          landedTime = t_elapsed;
        } else {
          // Bounce: lose 40% energy
          vy = Math.abs(vy) * 0.4;
          vx *= 0.5;
          vz *= 0.5;
        }
      }
    }

    // Phase 3: after landing, shrink away
    if (landed) {
      const t = Math.min((t_elapsed - landedTime) / 0.4, 1);
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
