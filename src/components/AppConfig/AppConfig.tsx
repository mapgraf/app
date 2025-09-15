import React, {ChangeEvent, useEffect, useMemo, useState} from 'react';
import { lastValueFrom } from 'rxjs';
import { css } from '@emotion/css';
import { AppPluginMeta, GrafanaTheme2, PluginConfigPageProps, PluginMeta } from '@grafana/data';
import { getBackendSrv } from '@grafana/runtime';
import { Button, Field, FieldSet, Input, SecretInput, useStyles2 } from '@grafana/ui';
import { testIds } from '../testIds';
import {convertUnixTimeToReadable} from "../../util";

export type AppPluginSettings = {
  apiPort: string;
};

type State = {
  apiPort: string;
  isApiTokenSet: boolean;
  apiToken: string;
};

export interface AppConfigProps extends PluginConfigPageProps<AppPluginMeta<AppPluginSettings>> {}

export const AppConfig = ({ plugin }: AppConfigProps) => {
  const s = useStyles2(getStyles);
  const { enabled, pinned, jsonData, secureJsonFields } = plugin.meta;
  const [state, setState] = useState<State>({
    apiPort: jsonData?.apiPort || "8089",
    apiToken: '',
    isApiTokenSet: Boolean(secureJsonFields?.apiToken),
  });

  const onResetApiToken = () =>
    setState({
      ...state,
      apiToken: '',
      isApiTokenSet: false,
    });

  const onChange = (event: ChangeEvent<HTMLInputElement>) => {
    setState({
      ...state,
      [event.target.name]: event.target.value.trim(),
    });
  };

    const [licInfo, setLicInfo] = useState<any>()

    useEffect(()=> {
        const getLicInfo = async ()=> {

            const location = window.location
            const host = location.host

            const observable = getBackendSrv()
                .fetch({
                    method: 'POST',
                    url: `/api/plugins/vaduga-mapgl-app/resources/ping`,
                    data: JSON.stringify({host}),
                    showErrorAlert: false
                })

            const res = await lastValueFrom(observable);
            const ping: any = res.data

            // const ping = await performPing(pingUrl)
            if (ping) {
                const {org, host, isPower, expiresAt} = ping
                console.log('ping', ping)

                const expiresFmt = convertUnixTimeToReadable(expiresAt * 1) // convert to number

                setLicInfo({
                    org, expiresAt, expiresFmt, host, isPower
                })
            }

        }
        getLicInfo()



    }, [])

    const time = Date.now();

    const hasExpired = time/1000 > licInfo?.expiresAt


    return (
    <div data-testid={testIds.appConfig.container}>
      <FieldSet label="License Settings">
          {/*description="Organization JWT token"*/}
        <Field label="JWT token" >
          <SecretInput
            width={60}
            data-testid={testIds.appConfig.apiToken}
            name="apiToken"
            value={state.apiToken}
            isConfigured={state.isApiTokenSet}
            placeholder={'Your JWT token'}
            onChange={onChange}
            onReset={onResetApiToken}
          />
        </Field>
          {licInfo && <div>
              licensed for: {licInfo?.org+' , '}
              host: {licInfo?.host}
              <p style={hasExpired ? {color: "#d10a85"} : {}}>
                  token expir{hasExpired? 'ed' : 'es'}: {licInfo?.expiresFmt}
              </p>
              tier: {licInfo?.isPower ? 'Power' : 'Lite'}

          </div>}

        {/*<Field label="Port" description="Port to start WebRTC signaling server for Mapgl in-browser-DB replication" className={s.marginTop}>*/}
        {/*  <Input*/}
        {/*    width={60}*/}
        {/*    name="apiPort"*/}
        {/*    data-testid={testIds.appConfig.apiPort}*/}
        {/*    value={state.apiPort}*/}
        {/*    placeholder={`E.g.: 8089`}*/}
        {/*    onChange={onChange}*/}
        {/*  />*/}
        {/*</Field>*/}

        <div className={s.marginTop}>
          <Button
            type="submit"
            data-testid={testIds.appConfig.submit}
            onClick={() =>
              updatePluginAndReload(plugin.meta.id, {
                enabled,
                pinned,
                jsonData: {
                  apiPort: state.apiPort,
                },
                // This cannot be queried later by the frontend.
                // We don't want to override it in case it was set previously and left untouched now.
                secureJsonData: state.isApiTokenSet
                  ? undefined
                  : {
                      apiToken: state.apiToken,
                    },
              })
            }
            disabled={Boolean(!state.apiPort || (!state.isApiTokenSet && !state.apiToken))}
          >
            Save API settings
          </Button>
        </div>
      </FieldSet>
    </div>
  );
};

const getStyles = (theme: GrafanaTheme2) => ({
  colorWeak: css`
    color: ${theme.colors.text.secondary};
  `,
  marginTop: css`
    margin-top: ${theme.spacing(3)};
  `,
});

const updatePluginAndReload = async (pluginId: string, data: Partial<PluginMeta<AppPluginSettings>>) => {
  try {
    await updatePlugin(pluginId, data);

    // Reloading the page as the changes made here wouldn't be propagated to the actual plugin otherwise.
    // This is not ideal, however unfortunately currently there is no supported way for updating the plugin state.
    window.location.reload();
  } catch (e) {
    console.error('Error while updating the plugin', e);
  }
};

export const updatePlugin = async (pluginId: string, data: Partial<PluginMeta>) => {
  const response = await getBackendSrv().fetch({
    url: `/api/plugins/${pluginId}/settings`,
    method: 'POST',
    data,
  });

  // @ts-ignore
  return lastValueFrom(response);
};
