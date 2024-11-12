<script>
	import { useThrelte, useTask } from "@threlte/core";
	import { onMount } from "svelte";
	import {
		EffectComposer,
		PixelationEffect,
		EffectPass,
		RenderPass,
		SMAAEffect,
		SMAAPreset,
		BloomEffect,
		KernelSize,
	} from "postprocessing";
	import Scene from "./Scene.svelte";
	import * as THREE from "three";

	const { scene, renderer, camera, size } = useThrelte();

	// renderer.setClearColor(0x4287f5, 0);
	scene.background = new THREE.Color( 0xa7e5f9 );

	// Adapt the default WebGLRenderer: https://github.com/pmndrs/postprocessing#usage
	const composer = new EffectComposer(renderer);

	const setupEffectComposer = (/** @type {import("three").Camera | undefined} */ camera) => {
		composer.removeAllPasses();
		composer.addPass(new RenderPass(scene, camera));
        composer.addPass(new EffectPass(camera, new PixelationEffect(0)));
		composer.addPass(
			new EffectPass(
				camera,
				new BloomEffect({
					intensity: 0.0,
					luminanceThreshold: 0.15,
					height: 512,
					width: 512,
					luminanceSmoothing: 0.08,
					mipmapBlur: true,
					kernelSize: KernelSize.MEDIUM,
				})
			)
		);

        // Anti-aliasing
		// composer.addPass(
		// 	new EffectPass(
		// 		camera,
		// 		new SMAAEffect({
		// 			preset: SMAAPreset.LOW,
		// 		})
		// 	)
		// );
	};

	// We need to set up the passes according to the camera in use
	$: setupEffectComposer($camera);
	$: composer.setSize($size.width, $size.height);

	const { renderStage, autoRender } = useThrelte();

	// We need to disable auto rendering as soon as this component is
	// mounted and restore the previous state when it is unmounted.
	onMount(() => {
		let before = autoRender.current;
		autoRender.set(false);
		return () => autoRender.set(before);
	});

	useTask(
		(delta) => {
			composer.render(delta);
		},
		{ stage: renderStage, autoInvalidate: false }
	);
</script>


