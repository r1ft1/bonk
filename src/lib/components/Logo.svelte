<script lang="ts">
	import { Text3DGeometry, Suspense, Float } from "@threlte/extras";
	import { T } from "@threlte/core";

	export let text
	export let position
</script>

<Float floatIntensity={2} speed={[1.5,1.5,1.5]} rotationSpeed={[1.5,1.5,1.5]} rotationIntensity={1.5}>
	<Suspense>
		<T.Mesh position={position}>
			<Text3DGeometry
				font={"/cherry-bomb.json"}
				text={text}
				size={1}
				height={0.2}
			/>
			<T.MeshStandardMaterial />
		</T.Mesh>
	</Suspense>
</Float>
