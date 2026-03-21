<script lang="ts">
  import { T } from "@threlte/core";
  import { Outlines, Edges, useGltf } from "@threlte/extras";

  let {
    position = [0, 0, 0],
    rotation = 0,
    scale = 1,
    kittens = 0,
    cats = 0,
    color = "orange",
  }: {
    position?: [number, number, number];
    rotation?: number;
    scale?: number;
    kittens?: number;
    cats?: number;
    color?: string;
  } = $props();

  const kittenGltf = useGltf("/kitten.glb");
  const catGltf = useGltf("/cat.glb");

  // Arrange pieces in a grid inside the box (2 rows x 4 cols)
  // Box body is 1.8x1.8, so space pieces within ~1.4 range
  function gridPos(index: number): [number, number, number] {
    const col = index % 4;
    const row = Math.floor(index / 4);
    const x = -0.45 + col * 0.3;
    const z = -0.15 + row * 0.3;
    const y = -0.5; // sit on bottom of box
    return [x, y, z];
  }
</script>

<T.Group {position} rotation.y={rotation} scale={[scale, scale, scale]}>
  <!-- Box walls (hollow — no top) -->
  <!-- Bottom -->
  <T.Mesh position={[0, -0.8, 0]}>
    <T.BoxGeometry args={[1.8, 0.06, 1.8]} />
    <T.MeshStandardMaterial color="#c4956a" />
  </T.Mesh>
  <!-- Front wall -->
  <T.Mesh position={[0, 0, 0.87]}>
    <T.BoxGeometry args={[1.8, 1.6, 0.06]} />
    <T.MeshStandardMaterial color="#c4956a" />
    <Outlines color="black" thickness={0.02} />
  </T.Mesh>
  <!-- Back wall -->
  <T.Mesh position={[0, 0, -0.87]}>
    <T.BoxGeometry args={[1.8, 1.6, 0.06]} />
    <T.MeshStandardMaterial color="#c4956a" />
    <Outlines color="black" thickness={0.02} />
  </T.Mesh>
  <!-- Left wall -->
  <T.Mesh position={[-0.87, 0, 0]}>
    <T.BoxGeometry args={[0.06, 1.6, 1.8]} />
    <T.MeshStandardMaterial color="#c4956a" />
    <Outlines color="black" thickness={0.02} />
  </T.Mesh>
  <!-- Right wall -->
  <T.Mesh position={[0.87, 0, 0]}>
    <T.BoxGeometry args={[0.06, 1.6, 1.8]} />
    <T.MeshStandardMaterial color="#c4956a" />
    <Outlines color="black" thickness={0.02} />
  </T.Mesh>
  <!-- Flap back -->
  <T.Mesh position={[0, 0.8, -0.8]} rotation.x={-0.5}>
    <T.BoxGeometry args={[1.7, 0.06, 0.85]} />
    <T.MeshStandardMaterial color="#d4a574" />
    <Outlines color="black" thickness={0.015} />
  </T.Mesh>
  <!-- Flap front -->
  <T.Mesh position={[0, 0.8, 0.8]} rotation.x={0.3}>
    <T.BoxGeometry args={[1.7, 0.06, 0.85]} />
    <T.MeshStandardMaterial color="#d4a574" />
    <Outlines color="black" thickness={0.015} />
  </T.Mesh>
  <!-- Flap left -->
  <T.Mesh position={[-0.8, 0.8, 0]} rotation.z={0.4}>
    <T.BoxGeometry args={[0.06, 0.85, 1.7]} />
    <T.MeshStandardMaterial color="#d4a574" />
    <Outlines color="black" thickness={0.015} />
  </T.Mesh>
  <!-- Flap right -->
  <T.Mesh position={[0.8, 0.8, 0]} rotation.z={-0.6}>
    <T.BoxGeometry args={[0.06, 0.85, 1.7]} />
    <T.MeshStandardMaterial color="#d4a574" />
    <Outlines color="black" thickness={0.015} />
  </T.Mesh>

  <!-- Kitten pieces inside box -->
  {#await kittenGltf then gltf}
    {#each Array(kittens) as _, i}
      {@const pos = gridPos(i)}
      <T.Mesh
        geometry={gltf.nodes.Kitten.geometry}
        position={pos}
        scale={[0.25, 0.25, 0.25]}
      >
        <T.MeshStandardMaterial {color} />
        <Outlines color="black" thickness={0.04} />
      </T.Mesh>
    {/each}
  {/await}

  <!-- Cat pieces inside box (placed after kittens in grid) -->
  {#await catGltf then gltf}
    {#each Array(cats) as _, i}
      {@const pos = gridPos(kittens + i)}
      <T.Mesh
        geometry={gltf.nodes.Cube.geometry}
        position={pos}
        scale={[0.25, 0.25, 0.25]}
      >
        <T.MeshStandardMaterial {color} />
        <Outlines color="black" thickness={0.04} />
        <Edges color="black" />
      </T.Mesh>
    {/each}
  {/await}
</T.Group>
