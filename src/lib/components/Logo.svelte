<script lang="ts">
	import { Text3DGeometry, Suspense, Float } from "@threlte/extras";
	import { T } from "@threlte/core";
</script>

<Float floatIntensity={2} speed={[1.5,1.5,1.5]} rotationSpeed={[1.5,1.5,1.5]} rotationIntensity={1.5}>
	<Suspense>
		<T.Mesh position={[0, 4, 0]}>
			<Text3DGeometry
				font={"/cherry-bomb.json"}
				text={"boop!"}
				size={1}
				height={0.2}
			/>
			<T.MeshStandardMaterial />
		</T.Mesh>
	</Suspense>
</Float>
