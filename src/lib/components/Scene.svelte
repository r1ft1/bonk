<script lang="ts">
	import { T } from "@threlte/core";
	import CameraControls from './CameraControls/CameraControls.svelte'
	import { cameraControls, mesh } from './CameraControls/stores'
	import { AmbientLight, DirectionalLight } from "three";
</script>

<T.PerspectiveCamera
	makeDefault
	position={[10, 10, 10]}
	on:create={({ ref }) => {
		ref.lookAt(0, 1, 0);
	}}
>
	<CameraControls
		on:create={({ ref }) => {
			$cameraControls = ref;
		}}
	/>
</T.PerspectiveCamera>

<T.DirectionalLight position={[0, 10, 0]} intensity={1} />
<T.AmbientLight intensity={0.5} />

<T.Mesh>
	<T.BoxGeometry args={[6, 1, 6]} />
	<T.MeshStandardMaterial color="lightblue" />
</T.Mesh>
