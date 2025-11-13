import { Dispatch, SetStateAction } from "react";

// A reusable empathetic error handler
export const handleError = (
  err: any,
  setError: Dispatch<SetStateAction<string>>,
  context: string = ""
) => {
  console.error(`[${context}]`, err);

  let message = "Hmm… something didn’t work as expected. Take a deep breath and try again. If the issue persists, refreshing the page might help.";

  if (err instanceof DOMException) {
    message = "It seems we don’t have permission to access your microphone or camera. Please allow access and try again — we know permissions can be tricky.";
  } 
  else if (err.message?.includes("No media stream")) {
    message = "We couldn’t access your microphone or screen. Take a moment to check your device settings, then try again. It happens to everyone!";
  } 
  else if (err.message?.includes("Network")) {
    message = "It looks like your internet connection dropped. Please check your connection and try again. You’ve got this!";
  } 
  else if (err.message?.includes("Upload failed")) {
    message = "Uh-oh… we had trouble uploading your file. Try again in a moment, and make sure your internet connection is stable.";
  } 
  else if (err.message?.includes("Processing failed")) {
    message = "Something went wrong while processing your file. Don’t worry — you can try again and it should work fine.";
  } 
  else if (typeof err === "string") {
    message = `Hmm… ${err}. Take a short pause, then try again.`;
  } 
  else if (err?.message) {
    message = `Oops… ${err.message}. Relax, and try the action once more — it usually works the second time!`;
  }

  setError(message);
};
