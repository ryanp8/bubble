import * as React from "react";
import * as Linking from "expo-linking";
import EventSource from "react-native-sse";
import { StyleSheet, View, Text, Button } from "react-native";
import { GestureHandlerRootView } from "react-native-gesture-handler";
import { NavigationContainer } from "@react-navigation/native";
import { createNativeStackNavigator } from "@react-navigation/native-stack";
import { DefaultTheme } from "@react-navigation/native";

import Home from "./screens/Home";
import Room from "./screens/Room";
import { StatusBar } from "expo-status-bar";

global.EventSource = EventSource;

const prefix = Linking.createURL("/");

const Stack = createNativeStackNavigator();
const theme = {
  ...DefaultTheme,
  colors: {
    ...DefaultTheme.colors,
    background: '#eeeeee'
  }
}

export default function App() {
  const linking = {
    prefixes: [prefix],
    config: {
      screens: {
        Home: "/",
        Room: "rooms/:id/:owner",
      },
    },
  };
  return (
    <GestureHandlerRootView style={{ flex: 1 }}>
      <StatusBar />
      <NavigationContainer linking={linking} fallback={<Text>hello</Text>} theme={theme}>
        <Stack.Navigator>
          <Stack.Screen name="Home" component={Home} />
          <Stack.Screen
            name="Room"
            component={Room}
            options={({ navigation, route }) => ({
              title: `${route.params.owner}'s Room`,
              headerLeft: () => (
                <Button
                  title="Home"
                  onPress={() => {
                    navigation.navigate("Home");
                  }}
                />
              ),
            })}
          />
        </Stack.Navigator>
      </NavigationContainer>
    </GestureHandlerRootView>
  );
}

const styles = StyleSheet.create({});
