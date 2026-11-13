import os
from PIL import Image

frames_dir = r"B:\workgit\llmmeter\docs\images\frames"
output_gif = r"B:\workgit\llmmeter\docs\images\llmmeter_demo.gif"

frame_files = [
    "frame_01.png",
    "frame_02.png",
    "frame_03.png",
    "frame_04.png",
    "frame_05.png",
    "frame_06.png",
]

images = []
# Durations in milliseconds for each frame
durations = [2200, 2200, 2500, 2200, 2200, 2200]

for f in frame_files:
    p = os.path.join(frames_dir, f)
    if os.path.exists(p):
        img = Image.open(p)
        # Resize slightly to 1024x576 for optimal GIF size under 3MB while maintaining crisp readability
        img = img.resize((1024, 576), Image.Resampling.LANCZOS)
        # Quantize with adaptive palette
        img = img.convert("P", palette=Image.Palette.ADAPTIVE, colors=256)
        images.append(img)

if images:
    images[0].save(
        output_gif,
        save_all=True,
        append_images=images[1:],
        duration=durations,
        loop=0,
        optimize=True,
    )
    size_mb = os.path.getsize(output_gif) / (1024 * 1024)
    print(f"✓ Generated {output_gif} ({size_mb:.2f} MB)")
else:
    print("Error: No frames found")
