# TopTen Image Tools — User Guide

This guide walks you through everything you need to know to convert images using TopTen Image Tools. No technical knowledge is needed.

---

## What does this app do?

TopTen Image Tools prepares your images so they are ready to upload to the CMS. It does two things automatically:

1. **Converts** your images to the right file format (JPG or PNG).
2. **Resizes** them so they are never wider or taller than 1 200 pixels — the maximum size the CMS needs. If your image is already small enough, it will not be changed in size.

Your original files are **never modified or deleted**. The app always creates new converted copies.

---

## Opening the app

| Platform | How to open |
|---|---|
| **macOS** | Double-click **TopTen Image Tools** in your Applications folder (or wherever you saved it) |
| **Windows** | Double-click **topten-image-tools.exe** |
| **Linux** | Double-click the file or run it from your file manager |

When the app opens you will see the **home screen** with three options.

---

## Step 1 — Choose what you want to convert

![Home screen — three mode cards](./images/home-screen.png)

Pick the option that matches your situation:

---

### 🖼 Single Image

Use this when you have **one image** to convert.

1. Click **Select**.
2. A file browser opens — navigate to your image and click **Open**.
3. The file name appears in the list. Click **Next →** to continue.

---

### 🗂 Multiple Images

Use this when you have **a handful of images** from different folders, or when you want to pick and choose which files to include.

1. Click **Select**.
2. Click **Add Image** and pick your first image.
3. Click **Add Image** again for each additional image you want to add. Build up the list one file at a time.
4. Once all your images are listed, click **Next →** to continue.

> **Tip:** If you accidentally add the wrong file, click **Clear** to start the list over.

---

### 📁 Entire Folder

Use this when you want to convert **all images inside one folder** at the same time.

1. Click **Select**.
2. A folder browser opens — navigate to the folder that contains your images and click **Open**.
3. The app will list every image it found. Click **Next →** to continue.

> **Note:** Only the top-level images in the folder are included. Images inside sub-folders are not processed.

---

## Step 2 — Choose the right format

The app asks you a short question to work out whether your images should be saved as **JPG** or **PNG**.

### What is the difference?

| Format | Best for |
|---|---|
| **JPG** | Photos, landscapes, people, product shots, gradients. Smaller file size. |
| **PNG** | Logos, screenshots, infographics, anything with text on it, or images with a transparent background. Crisp and lossless. |

### How to answer

Read the four options and click **Select** next to the one that best describes the images you are converting:

| Option | Examples |
|---|---|
| 📷 Photos or natural images | Holiday photos, product photography, lifestyle shots |
| ✏️ Graphics with text or logos | Infographics, screenshots, team headshots with name overlays |
| 🔍 Images with a transparent background | Icons, cutout images, product images with no background |
| 📢 Website hero banners / featured images | The large image at the top of an article or homepage section |

**If you picked "Hero banners"**, one more question appears:

> *Do your banners contain text overlays or logos?*

- Click **Yes** if the banner has a title, headline, or logo placed on top of the image.
- Click **No** if it is a pure photograph with no text.

---

### The recommendation card

After you answer, a coloured card appears with the recommended format and a short reason explaining why. For example:

> **PNG recommended ✏️**
> Images with text, logos, or sharp lines preserve quality best as PNG (lossless).

If you disagree with the recommendation, you can override it by clicking **Use JPG instead** or **Use PNG instead** before moving on.

When you are happy with the format, click **Use this format →**.

---

## Step 3 — Choose where to save the converted files

The app suggests saving the converted images in the **same folder as your original files** — this is usually the most convenient option.

- If that is fine, just click **Convert Now**.
- If you want to save them somewhere else (for example, a dedicated "CMS uploads" folder), click **Browse…**, navigate to the folder you want, and then click **Convert Now**.

> **Tip:** Saving to a separate folder keeps your originals and converted files neatly apart and makes it easy to find what to upload.

---

## Step 4 — Conversion in progress

A progress bar shows how far along the conversion is.

- Each file is shown by name as it is processed.
- If something goes wrong with a single file (for example, the file is corrupted), a warning appears for that file but the rest will still be converted.
- If you need to stop, click **Cancel** — no files will be left in a broken state.

---

## Step 5 — Results

When the conversion is complete you will see a summary screen showing:

- How many images were converted successfully.
- How much storage space was saved compared to the originals (saving space is normal when converting to JPG; some formats like PNG can occasionally be slightly larger — both are fine).
- Any files that could not be converted, with a short error message for each.

Click **Open Output Folder** to jump straight to the folder containing your converted files — they are ready to upload to the CMS.

Click **Convert More Images** to go back to the home screen and start another batch.

---

## Tips & common questions

**Can I convert a mix of photos and graphics at the same time?**
The wizard picks one format for the whole batch. If your batch contains a true mix (some photos, some graphics), split them into two separate conversions — one set converted to JPG and the other to PNG.

**Will my originals be overwritten?**
No. The app always creates new files. Your originals are never touched.

**What if a converted file already exists in the output folder?**
The app will not overwrite it. It automatically adds a number to the new file name (e.g. `banner_1.jpg`) so nothing is lost.

**What image files does the app accept?**
JPG, JPEG, PNG, GIF, BMP, TIFF, TIF, and WebP.

**My image is already smaller than 1 200 px — will it be changed?**
No. Images that are already within the 1 200 px limit are not resized.

**The app says it converted to JPG but the file size is larger than the original — is that a problem?**
This can happen when converting a very small or heavily compressed JPG to a slightly higher-quality setting. The file is still perfectly valid for the CMS. If file size is a concern, you can re-run the conversion and choose PNG, or check with your CMS administrator.

---

## Need help?

Contact your team's CMS administrator or open an issue at the project's GitHub repository.
