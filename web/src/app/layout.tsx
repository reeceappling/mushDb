import type { Metadata } from "next";
import "./global.css"; // Also includes postcss/tailwind
import "@/app/ui/global.css"
import "@/app/ui/mostRecentImage.css"
import "@/app/ui/flexedAreas.css"
import "@/app/ui/subform.css"
import "@/app/ui/topBar.css"
import "@/app/ui/onViewCreators.css"
import "@/app/ui/transferDisplay.css"
import "@/app/ui/listPage.css"
import "@/app/ui/project.css"
import { Geist, Geist_Mono } from "next/font/google";
import {GoogleAnalytics} from "@next/third-parties/google" // TODO: ensure works
import {BaseExternalUrl, GoogleAnalyticsId} from "@/app/components/Constants";
import { CookieManager } from "react-cookie-manager";
import {ReaderOptionsContextProvider} from "@/app/components/formSubcomponents/readerWriterButtons/readerOptsContext";
import {GetReaderWriterNames} from "@/app/components/serverActions";

// TODO: Cache Components adoption. Refactor this route so this opt-out can be removed.
// See: https://nextjs.org/docs/app/guides/migrating-to-cache-components
//export const instant = false;

const geistSans = Geist({
  variable: "--font-geist-sans",
  subsets: ["latin"],
});

const geistMono = Geist_Mono({
  variable: "--font-geist-mono",
  subsets: ["latin"],
});

export const metadata: Metadata = {
  title: {
    template:  '%s | MushDB',
    default: "MushDB",
  },
  description: "Mycology Database",
  generator: `Wouldn't you like to know, weather boy`,
  applicationName: 'MushDb',
  //referrer: 'origin-when-cross-origin', // TODO: ok?
  keywords: ['MushDb', 'ReeceAppling', 'Reece Appling','Reece','Appling',"mycology","mushrooms","fungi"],
  authors: [{ name: 'Reece Appling', url: 'reece.appli.ng' }], // TODO: ok?
  creator: 'Reece Appling', // TODO: ok?
  publisher: 'Reece Appling', // TODO: ok?
  category: 'technology', // TODO: FIX!
  // formatDetection: { // TODO: ok?
  //   email: false,
  //   address: false,
  //   telephone: false,
  // },
  // openGraph: { // TODO: FIX THIS ON ALL!
  //   title: 'Next.js',
  //   description: 'The React Framework for the Web',
  //   type: 'article',
  //   publishedTime: '2023-01-01T00:00:00.000Z', // TODO: FIX
  //   authors: ['Seb', 'Josh'],
  // },
  // robots: { // TODO: ???
  //   index: true,
  //   follow: true,
  //   nocache: false,
  //   googleBot: {
  //     index: true,
  //     follow: true,
  //     noimageindex: false,
  //     'max-video-preview': -1,
  //     'max-image-preview': 'large',
  //     'max-snippet': -1,
  //   },
  // },
  // icons: { // TODO: ??? icon came from https://www.favicon.cc/?action=icon&file_id=146503
  //   icon: '/icon.png',
  //   shortcut: '/shortcut-icon.png',
  //   apple: '/apple-icon.png',
  //   other: {
  //     rel: 'apple-touch-icon-precomposed',
  //     url: '/apple-touch-icon-precomposed.png',
  //   },
  // },
  //manifest: 'https://nextjs.org/manifest.json', // TODO: ???
  // twitter: { // TODO: ????
  //   card: 'summary_large_image',
  //   title: 'Next.js',
  //   description: 'The React Framework for the Web',
  //   siteId: '1467726470533754880',
  //   creator: '@nextjs',
  //   creatorId: '1467726470533754880',
  //   images: ['https://nextjs.org/og.png'], // Must be an absolute URL
  // },
  // verification: { // TODO: ???
  //   google: 'google',
  //   yandex: 'yandex',
  //   yahoo: 'yahoo',
  //   other: {
  //     me: ['my-email', 'my-link'],
  //   },
  // },
  // itunes: { // TODO: ???
  //   appId: 'myAppStoreID',
  //   appArgument: 'myAppArgument',
  // },
  // appleWebApp: { // TODO: ???
  //   title: 'Apple Web App',
  //   statusBarStyle: 'black-translucent',
  //   startupImage: [
  //     '/assets/startup/apple-touch-startup-image-768x1004.png',
  //     {
  //       url: '/assets/startup/apple-touch-startup-image-1536x2008.png',
  //       media: '(device-width: 768px) and (device-height: 1024px)',
  //     },
  //   ],
  // },
  alternates: { // TODO: this!
    // canonical: 'https://nextjs.org', // TODO: ???
    languages: {
      'en-US': 'https://nextjs.org/en-US',
      // 'de-DE': 'https://nextjs.org/de-DE', // TODO: ???
    },
    // media: { // TODO: fix!
    //   'only screen and (max-width: 600px)': 'https://nextjs.org/mobile',
    // },
    // types: { // TODO: fix!
    //   'application/rss+xml': 'https://nextjs.org/rss',
    // },
  },
  // appLinks: { // TODO: ???
  //   ios: {
  //     url: 'https://nextjs.org/ios',
  //     app_store_id: 'app_store_id',
  //   },
  //   android: {
  //     package: 'com.example.android/package',
  //     app_name: 'app_name_android',
  //   },
  //   web: {
  //     url: 'https://nextjs.org/web',
  //     should_fallback: true,
  //   },
  // },
  // archives: ['https://nextjs.org/13'],
  // assets: ['https://nextjs.org/assets'],
  // bookmarks: ['https://nextjs.org/13'],
  // pinterest: {
  //   richPin: true,
  // },
  // other: {
  //   custom: 'meta',
  // },
};
function GoogleAnalyticsComponent(){
  if (GoogleAnalyticsId=="none"){
    return null
  }
  return <GoogleAnalytics gaId={GoogleAnalyticsId} />
}
function GdprCookieManager({children}: {children: React.ReactNode}){
  return <CookieManager
      translations={{
        title: "What cookies and storage do you want to allow?",
        message:
            "We value your privacy. Choose which cookies you want to allow. Essential cookies are always enabled as they are necessary for the website to function properly.",
        buttonText: "Accept All",
        declineButtonText: "Decline Non-Essential", // TODO:
        manageButtonText: "Manage Cookies",
        privacyPolicyText: "Privacy Policy",
      }}
      categories={[
        // Override a built-in's copy (optional):
        { id: "Analytics", description: "Helps us improve the product" },
        { id: "Social", description: "Helps us improve the product" },
        { id: "Advertising", description: "Helps us improve the product" },
        // Add custom categories:
        { id: "FunctionalCookies", title: "Functional Cookies", description: "Remember your settings with cookies",defaultConsent: true, trackerDomains: ["ads.example.com", "track.partner.com"]},
        { id: "FunctionalSessionStorage", title: "Functional Cookies", description: "Remember your settings using short-term session storage",defaultConsent: true, trackerDomains: ["ads.example.com", "track.partner.com"]},
        { id: "FunctionalLocalStorage", title: "Functional Local Storage", description: "Remember your settings using long-term local storage",defaultConsent: true, trackerDomains: ["ads.example.com", "track.partner.com"]},
      ]}
      showManageButton={true}
      privacyPolicyUrl={BaseExternalUrl+"/policies/privacy"}
      theme="light"
      displayType="popup"
      onManage={(preferences) => {
        if (preferences) {
          console.log("Cookie preferences updated:", preferences);
        }
      }}
      onAccept={() => {
        console.log("User accepted all cookies");
        // Analytics tracking can be initialized here
      }}
      onDecline={() => {
        console.log("User declined all cookies");
        // Handle declined state if needed
      }}
  >
    {children}
  </CookieManager>
}
async function Providers({children}: {children: React.ReactNode}){
  const readers = await GetReaderWriterNames() // Done on the server
  return <GdprCookieManager>
    {/* TODO: user initial reader state based off of most recent?*/}
    <ReaderOptionsContextProvider initialState={{options: readers, selected: undefined}}>{/* TODO: if not work, reenable everywhere else and delete this*/}
      {children}
    </ReaderOptionsContextProvider>
  </GdprCookieManager>
}

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en">
    <body className={`${geistSans.variable} ${geistMono.variable}`}>
      <Providers>{children}</Providers>
    </body>
    <GoogleAnalyticsComponent/>
    </html>
  );
}
