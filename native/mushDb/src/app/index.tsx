import * as Device from 'expo-device';
import {Button, Platform, StyleSheet} from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';

import { AnimatedIcon } from '@/components/animated-icon';
import { HintRow } from '@/components/hint-row';
import { ThemedText } from '@/components/themed-text';
import { ThemedView } from '@/components/themed-view';
import { WebBadge } from '@/components/web-badge';
import { BottomTabInset, MaxContentWidth, Spacing } from '@/constants/theme';
import PhotoPicker from "@/components/ui/photos";
import React, {useEffect, useState} from "react";
import axios from 'axios';
import AsyncStorage from '@react-native-async-storage/async-storage';

// Home screen. No Information here, just gives the user the ability to log in!

function getDevMenuHint() {
  if (Platform.OS === 'web') {
    return <ThemedText type="small">use browser devtools</ThemedText>;
  }
  if (Device.isDevice) {
    return (
      <ThemedText type="small">
        shake device or press <ThemedText type="code">m</ThemedText> in terminal
      </ThemedText>
    );
  }
  const shortcut = Platform.OS === 'android' ? 'cmd+m (or ctrl+m)' : 'cmd+d';
  return (
    <ThemedText type="small">
      press <ThemedText type="code">{shortcut}</ThemedText>
    </ThemedText>
  );
}

export default function HomeScreen() {
  const [loggedIn, setLoggedIn] = useState<boolean>(false);
  useEffect(()=>{
    isLoggedIn().then(setLoggedIn).catch((e)=>{
      console.error("FAILED TO LOGIN: "+JSON.stringify(e));
    })
    // TODO: FIX THIS SO IT CHECKS THAT USER IS LOGGED IN!
  },[])
  return (
    <ThemedView style={styles.container}>
      <SafeAreaView style={styles.safeArea}>
        <ThemedView style={styles.heroSection}>
          <AnimatedIcon />
          <ThemedText type="title" style={styles.title}>
            Hello World!
          </ThemedText>
        </ThemedView>

        <ThemedText type="code" style={styles.code}>
          {loggedIn ? <Button title="Log Out" onPress={()=>{
            logout().then(()=>{
              setLoggedIn(false);
            }).catch((e)=>{
              // TOOD: this!
            })
          }} /> : <Button title="Login" onPress={()=>{
            login().then((loggedIn)=>{
              setLoggedIn(loggedIn);
            }).catch((e)=>{
              console.error("FAILED TO LOGIN 2: "+JSON.stringify(e));
              // TODO: this!
            })
          }} />}
        </ThemedText>

        <ThemedView type="backgroundElement" style={styles.stepContainer}>
          <HintRow
            title="Try editing"
            hint={<ThemedText type="code">src/app/index.tsx</ThemedText>}
          />
          <HintRow title="Dev tools" hint={getDevMenuHint()} />
          <HintRow
            title="Fresh start"
            hint={<ThemedText type="code">npm run reset-project</ThemedText>}
          />
        </ThemedView>
        <PhotoPicker></PhotoPicker>

        {Platform.OS === 'web' && <WebBadge />}
      </SafeAreaView>
    </ThemedView>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    justifyContent: 'center',
    flexDirection: 'row',
  },
  safeArea: {
    flex: 1,
    paddingHorizontal: Spacing.four,
    alignItems: 'center',
    gap: Spacing.three,
    paddingBottom: BottomTabInset + Spacing.three,
    maxWidth: MaxContentWidth,
  },
  heroSection: {
    alignItems: 'center',
    justifyContent: 'center',
    flex: 1,
    paddingHorizontal: Spacing.four,
    gap: Spacing.four,
  },
  title: {
    textAlign: 'center',
  },
  code: {
    textTransform: 'uppercase',
  },
  stepContainer: {
    gap: Spacing.three,
    alignSelf: 'stretch',
    paddingHorizontal: Spacing.three,
    paddingVertical: Spacing.four,
    borderRadius: Spacing.four,
  },
});

const cfg = {
  baseURL: 'https://mush.appli.ng', // TODO: FIX THIS!
  domain: 'mush.appli.ng', // TODO: FIX THIS!
};
const api = axios.create({
  baseURL: cfg.baseURL,
  withCredentials: true, // ensures that Axios sends cookies with requests
});
api.interceptors.request.use(async (config) => {
  const sessionCookieString = await AsyncStorage.getItem(gothicSessionCookieKey);
  if (sessionCookieString) {
    config.headers[gothicSessionCookieKey] = sessionCookieString;
  }
  return config;
});
const gothicSessionCookieKey = "_gothic_session"
async function login():Promise<boolean>{
  const loggedIn = await fakeLogin() // TODO: REMOVE! THIS IS FOR TESTING ONLY!
  return loggedIn
  //return false // TODO: success should return true
  // TODO: DO THIS!
}
async function fakeLogin():Promise<boolean>{
  await AsyncStorage.setItem(gothicSessionCookieKey, "fakeSession"); // TODO: remove later
  return true // TODO: success should return true
}
async function ParseLoginResponse(response: Response){ // TODO: USE THIS!
  const setCookieHeader = response.headers.get('set-cookie');
  const match = setCookieHeader?.match(/_gothic_session=([^;]+)/);

  if (match) {
    const sessionId = match[1];
    await AsyncStorage.setItem(gothicSessionCookieKey, sessionId);
  }
}
async function isLoggedIn():Promise<boolean>{ // TODO: DO THIS!
  try {
    const url = cfg.baseURL;
    setTimeout(()=>{}, 3000);
    // TODO: send request to server and check response
    return true
  } catch (error) {
    return false;
  }
}
async function logout(){ // TODO: DO THIS!
  try {
    // TODO: actually log out
    // Update local storage

    // TODO: ???
  } catch (error) {
    // TODO: ???
  }
  try {
    await AsyncStorage.removeItem(gothicSessionCookieKey);
    // TODO: ???
  } catch (error) {
    // TODO: ???
  }
  return
}


