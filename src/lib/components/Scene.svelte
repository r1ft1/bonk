<script lang="ts">
	import { T, useThrelte, useTask } from "@threlte/core";
	import CameraControls from "./CameraControls/CameraControls.svelte";
	import { cameraControls, mesh } from "./CameraControls/stores";
	import Board from "./Board.svelte";
	import Logo from "./Logo.svelte";
	import { Stars } from "@threlte/extras";


	const { camera, renderMode,  } = useThrelte();
	renderMode.set("always");

</script>

<T.PerspectiveCamera
	makeDefault
	position={[10, 10, 10]}
	on:create={({ ref }) => {
		ref.lookAt(0, 1, 0);
		// camera = ref;
	}}
>
	<!-- <CameraControls
		on:create={({ ref }) => {
			$cameraControls = ref;
			console.log($cameraControls)
			console.log(ref)
		}}
	/> -->
</T.PerspectiveCamera>

<T.DirectionalLight position={[-1, 10, 0]} intensity={1} />
<T.AmbientLight intensity={0.7} />

<Logo/>

<Board />

<Stars />
