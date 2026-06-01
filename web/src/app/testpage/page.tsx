import React from "react";
import PageWrapper from "@/app/components/clientGeneric";
import TestSlider from "@/app/testpage/testClient";// TODO: del?
import {TransfersOutDisplay} from "@/app/components/transferClient";
import {TopPageHeaderLevel} from "@/app/components/Constants";
import PlateDisplay from "@/app/components/plateClient";
import {TestPlateOk} from "@/app/components/plateServer"; // TODO: del?
import {Closeable, TestNewAgarBatch, TestNewPlate} from "@/app/testpage/client";
import ProjectDisplay from "@/app/components/projectClient";
import {TestProjectOk, TestProjectOk2} from "@/app/components/projectServer";// TODO: del?
import {DisplayFormWrapper} from "@/app/components/common";

export default async function Page({
                                       params,
                                   }: {
    params: Promise<{}>,
}) {
    const {} = await params
    const onChange = (event: Event, value: number, activeThumb: number) => {
        value.toString()
    }
    return <div>
        {/*<Closeable title={"Text To Speech"}>*/}
        {/*    <PageWrapper props={{pageType:"testPage",readers:[]}}>*/}
        {/*        <div className={"fullPage"}>*/}
        {/*            <TextToSpeech />*/}
        {/*        </div>*/}
        {/*    </PageWrapper>*/}
        {/*</Closeable>*/}
        {/*<Closeable title={"Speech to text to speech"}>*/}
        {/*    <PageWrapper props={{pageType:"testPage",readers:[]}}>*/}
        {/*        <div className={"fullPage"}>*/}
        {/*            <SpeechToText/>*/}
        {/*        </div>*/}
        {/*    </PageWrapper>*/}
        {/*</Closeable>*/}
        <Closeable title={"CreatePlate"}>
            <PageWrapper props={{pageType:"testPage",readers:[]}}>
                <div className={"fullPage"}>
                    <TestNewPlate />
                </div>
            </PageWrapper>
        </Closeable>
        <Closeable title={"CreateAgarBatch"}>
            <PageWrapper props={{pageType:"testPage",readers:[]}}>
                <div className={"fullPage"}>
                    <TestNewAgarBatch />
                </div>
            </PageWrapper>
        </Closeable>
        <Closeable title={"PlateDisplay"}>
        <PageWrapper props={{pageType: "view", readers: ["reader1","reader2"]}}>
            <div className={"fullPage"}>
                <PlateDisplay data={TestPlateOk()} readonly={false} id={"testId"} isTopLevel={true}
                              headerLevel={TopPageHeaderLevel}/>
            </div>
        </PageWrapper>
        </Closeable>
        <Closeable title={"TestSlider"}>
        <div className={"whiteBackground"}>

            <TestSlider defaultValue={2}/>
        </div>
        </Closeable>
        <Closeable title={"TestTransfersOut"}>
        <PageWrapper props={{pageType:"testPage",readers:[]}}>
            <div className={"fullPage"}>
                <DisplayFormWrapper entryType={"plate"}>
                    <TransfersOutDisplay headerTxt={"Transfers"} thisId={"1"} thisEntryType={"plate"}
                                         transfersOut={["exTransfer1","exTransfer2"]}
                                         allowNewTransferCreation={true}/>
                </DisplayFormWrapper>

            </div>
        </PageWrapper>
        </Closeable>
        <Closeable title={"TestProjectWrite1"}>
            <PageWrapper props={{pageType:"testPage",readers:[]}}>
                <div className={"fullPage"}>
                    <ProjectDisplay data={TestProjectOk()} id={"TestProjectWrite"} readonly={false} isTopLevel={true} />
                </div>
            </PageWrapper>
        </Closeable>
        <Closeable title={"TestProjectWriteNoPermsNotComplete"}>
            <PageWrapper props={{pageType:"testPage",readers:[]}}>
                <div className={"fullPage"}>
                    <ProjectDisplay data={TestProjectOk2()} id={"TestProjectWrite"} readonly={false} isTopLevel={true} />
                </div>
            </PageWrapper>
        </Closeable>

    </div>

}
