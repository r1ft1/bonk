<script lang="ts">
	import { animConfig } from "./stores";

	let visible = $state(false);

	const defaults = {
		arcDuration: 0.6, arcHeight: 5, arcLandThreshold: 0.85,
		slideDuration: 0.25, slideArcHeight: 0.2,
		bumpDuration: 0.15, bumpDistance: 1.5, bumpArcHeight: 0.3,
		gravity: 18, bumpVelocityY: 3, bumpVelocityXZ: 4, groundY: -2.7,
		bounceEnergyLoss: 0.4, bounceFriction: 0.5, bounceMinVelocity: 1,
		tumbleSpeed: 6, shrinkDuration: 0.4,
	};

	const sliders: { key: string; label: string; tip: string; min: number; max: number; step: number }[] = [
		// Placement arc
		{ key: "arcDuration", label: "Arc Duration", tip: "How long the placement arc animation takes (seconds)", min: 0.1, max: 2, step: 0.05 },
		{ key: "arcHeight", label: "Arc Height", tip: "Peak height of the arc when placing a piece", min: 1, max: 15, step: 0.5 },
		{ key: "arcLandThreshold", label: "Arc Land Trigger %", tip: "At what % of the arc to trigger boop animations (0.85 = 85%)", min: 0.5, max: 1, step: 0.01 },

		// Slide
		{ key: "slideDuration", label: "Slide Duration", tip: "How long a booped piece takes to slide to its new position", min: 0.05, max: 1, step: 0.05 },
		{ key: "slideArcHeight", label: "Slide Arc Height", tip: "How high a sliding piece lifts off the board during its slide", min: 0, max: 1, step: 0.05 },

		// Flying
		{ key: "bumpDuration", label: "Bump Duration", tip: "Initial bump animation before the piece goes into freefall", min: 0.05, max: 0.5, step: 0.01 },
		{ key: "bumpDistance", label: "Bump Distance", tip: "How far the initial bump pushes the piece outward", min: 0.5, max: 4, step: 0.1 },
		{ key: "bumpArcHeight", label: "Bump Arc Height", tip: "How high the piece lifts during the initial bump", min: 0, max: 2, step: 0.05 },
		{ key: "gravity", label: "Gravity", tip: "Downward acceleration during freefall after being booped off", min: 5, max: 40, step: 1 },
		{ key: "bumpVelocityY", label: "Launch Up Speed", tip: "Upward velocity when the piece launches into freefall", min: 0, max: 10, step: 0.5 },
		{ key: "bumpVelocityXZ", label: "Launch Out Speed", tip: "Horizontal velocity when the piece launches into freefall", min: 0, max: 10, step: 0.5 },
		{ key: "groundY", label: "Ground Level", tip: "Y position of the ground plane for bounce detection", min: -5, max: 0, step: 0.1 },
		{ key: "bounceEnergyLoss", label: "Bounce Energy", tip: "How much velocity is kept after each bounce (0=dead stop, 1=perfect bounce)", min: 0, max: 1, step: 0.05 },
		{ key: "bounceFriction", label: "Bounce Friction", tip: "How much horizontal speed is lost on each bounce (0=none, 1=full stop)", min: 0, max: 1, step: 0.05 },
		{ key: "bounceMinVelocity", label: "Bounce Min Vel", tip: "Below this velocity the piece stops bouncing and settles", min: 0, max: 5, step: 0.1 },
		{ key: "tumbleSpeed", label: "Tumble Speed", tip: "How fast the piece rotates while flying through the air", min: 0, max: 15, step: 0.5 },
		{ key: "shrinkDuration", label: "Shrink Duration", tip: "How long the piece takes to shrink and disappear after landing", min: 0.1, max: 1, step: 0.05 },
	];

	function handleInput(key: string, value: number) {
		$animConfig = { ...$animConfig, [key]: value };
	}
</script>

<button class="toggle" onclick={() => visible = !visible}>
	{visible ? '\u2715' : '\u2699'}
</button>

{#if visible}
	<div class="panel">
		<h3>Animation Tuning</h3>
		<button class="dump" onclick={() => console.log(JSON.stringify($animConfig))}>Print to Console</button>
		{#each sliders as s}
			<label>
				<span class="label" title={s.tip}>{s.label}</span>
				<input
					type="range"
					min={s.min}
					max={s.max}
					step={s.step}
					value={$animConfig[s.key as keyof typeof $animConfig]}
					oninput={(e) => handleInput(s.key, parseFloat((e.target as HTMLInputElement).value))}
				/>
				<span class="value">{($animConfig[s.key as keyof typeof $animConfig]).toFixed(2)}</span>
				{#if $animConfig[s.key as keyof typeof $animConfig] !== defaults[s.key as keyof typeof defaults]}
					<button class="reset" title="Reset to default ({defaults[s.key as keyof typeof defaults]})" onclick={() => handleInput(s.key, defaults[s.key as keyof typeof defaults])}>↺</button>
				{/if}
			</label>
		{/each}
	</div>
{/if}

<style>
	.toggle {
		position: fixed;
		bottom: 0.5rem;
		right: 0.5rem;
		z-index: 1000;
		width: 2rem;
		height: 2rem;
		border-radius: 50%;
		border: 2px solid rgba(180, 160, 140, 0.4);
		background: rgba(250, 246, 240, 0.95);
		font-size: 1rem;
		cursor: pointer;
		display: flex;
		align-items: center;
		justify-content: center;
		pointer-events: all;
	}

	.panel {
		position: fixed;
		bottom: 3rem;
		right: 0.5rem;
		z-index: 1000;
		width: 280px;
		max-height: 80vh;
		overflow-y: auto;
		background: rgba(250, 246, 240, 0.95);
		border: 2px solid rgba(180, 160, 140, 0.3);
		border-radius: 12px;
		padding: 0.8rem;
		box-shadow: 0 4px 16px rgba(100, 80, 60, 0.15);
		pointer-events: all;
	}

	.dump {
		width: 100%;
		margin-bottom: 0.5rem;
		padding: 0.3rem;
		border: 1px solid rgba(180, 160, 140, 0.4);
		border-radius: 6px;
		background: rgba(180, 160, 140, 0.15);
		font-family: "Nunito", sans-serif;
		font-size: 0.65rem;
		font-weight: 600;
		color: #5a4a3a;
		cursor: pointer;
	}

	h3 {
		font-family: "Nunito", sans-serif;
		font-size: 0.85rem;
		font-weight: 700;
		color: #5a4a3a;
		margin: 0 0 0.5rem 0;
	}

	label {
		display: grid;
		grid-template-columns: 1fr 90px 40px 16px;
		align-items: center;
		gap: 0.3rem;
		margin-bottom: 0.3rem;
	}

	.reset {
		width: 16px;
		height: 16px;
		padding: 0;
		border: none;
		border-radius: 50%;
		background: rgba(180, 160, 140, 0.25);
		font-size: 0.55rem;
		line-height: 1;
		cursor: pointer;
		color: #5a4a3a;
		display: flex;
		align-items: center;
		justify-content: center;
	}

	.label {
		font-family: "Nunito", sans-serif;
		font-size: 0.65rem;
		font-weight: 600;
		color: #5a4a3a;
	}

	input[type="range"] {
		width: 100%;
		height: 4px;
		accent-color: #9a8a7a;
	}

	.value {
		font-family: "Nunito", sans-serif;
		font-size: 0.6rem;
		font-weight: 700;
		color: #9a8a7a;
		text-align: right;
	}
</style>
