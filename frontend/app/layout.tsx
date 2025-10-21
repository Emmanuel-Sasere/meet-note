import type { Metadata } from "next";
import { Geist, Geist_Mono } from "next/font/google";
import "./globals.css";
import { Analytics } from "@vercel/analytics/next";

const geistSans = Geist({
  variable: "--font-geist-sans",
  subsets: ["latin"],
});

const geistMono = Geist_Mono({
  variable: "--font-geist-mono",
  subsets: ["latin"],
});

export const metadata: Metadata = {
 title: 'Noted - Free AI Meeting Transcription & Summarization',
  description: 'Record and transcribe meetings, lectures, and audio with AI. Free online tool for accurate speech-to-text conversion and automatic summarization.',
  keywords: 'transcription, meeting notes, AI transcription, speech to text, audio transcription, video transcription, meeting summarization',
  openGraph: {
    title: 'Noted - AI Meeting Transcription',
    description: 'Free AI-powered meeting transcription and summarization tool',
    type: 'website',
  },
  twitter: {
    card: 'summary_large_image',
    title: 'Noted - AI Transcription',
    description: 'Free AI-powered meeting transcription',
  },
}

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
     <html lang="en" suppressHydrationWarning>
   
      <body
        className={`${geistSans.variable} ${geistMono.variable} antialiased`}
      suppressHydrationWarning>
        {children}
        <Analytics />
      </body>
    </html>
  );
}
