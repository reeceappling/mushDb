'use client'

import React, {JSX, useContext, useState} from "react";
import {IsValidNote, NewEntryNotes, Note, NotesFormArea} from "@/app/components/formSubcomponents/notes";
import {AddCreatedTriColFunction, AllEntries, OnViewCreatorQuadCol} from "@/app/components/formSubcomponents/shared";
import ID from "@/app/components/formSubcomponents/id";
import DateArea from "@/app/components/formSubcomponents/date";
import {
    CreatedLinkFor,
    DisplayFormWrapper,
    DisplayInput,
    DoCreateRequest,
    DoUpdateRequest,
    ExistingRecentSelector,
    FlexedArea,
    FlexedSinglesGroup,
    ListPageItems,
    ListPageTable,
    ListTableColumn,
    NewColumn,
    NewEntryFormWrapper,
    NewEntryInput,
    NumberToDateStr,
    OptionalArrayOfType,
    OptionalSimpleKey, RequiredKey
} from "@/app/components/common";
import {DisposedDisplay, ErrorDisplay} from "@/app/components/formSubcomponents/commonClient";
import {GrainBatchData} from "./grainBatchServer";
import {JarRecipeArea, JarRecipeSelector} from "@/app/components/jarRecipeClient";
import {NumericalArea} from "@/app/components/formSubcomponents/numericInput";
import {JarRecipeData} from "@/app/components/jarRecipeServer";
import {NewJarForm} from "@/app/components/jarClient";
import {JarData} from "@/app/components/jarServer";
import EntryLinkForId, {EntryLinkWrapper} from "@/app/components/formSubcomponents/entryLink";
import {OnViewCreatorsTriColArea} from "@/app/components/formSubcomponents/ovc";
import {InitialNotesState} from "@/app/components/formSubcomponents/initialState";
import {allCookies, CookiesContext} from "@/app/components/formSubcomponents/cookiesContext/cookies";
import {
    AclDisplay, MarshalAcl, UnmarshalAcl,
    TogglableAreaWithDepth, NewAllCanWriteAcl
} from "@/app/components/accessControlClient";
import { ACL } from "./accessControlServer";
import {ActionTypes, useModalContext} from "@/app/components/formSubcomponents/modalContext/modal";
import {GrainWaterJarData} from "@/app/components/grainWaterJarServer";
import {PcRunArea} from "@/app/components/pcRunClient";
import {GrainBatchArea} from "@/app/components/grainBatchClient";

// TODO: may want to link grainwater to agar batches. Notes on this left in the agar batch go struct.

export function AssertGrainWaterJar(input: any): asserts input is GrainWaterJarData {
    if (typeof input !== 'object') {
        throw new Error('Input is not an object! Input is ' + typeof input);
    }
    // required simple keys
    const requiredSimpleKeys = new Map<string, string>([
        ['_id', 'string'],
        ['grainBatch', 'string'],
        ['creationDate', 'number'],
        ['lastUpdated', 'number'],
    ])
    for (const [key, expType] of requiredSimpleKeys) {
        if (!(key in input && typeof input[key] === expType)) {
            throw new Error('Grain Water Jar assertion failure: ' + key + 'was not type ' + expType + '. Was ' + (typeof input[key]));
        }
    }
    // optional simple keys
    const optionalSimpleKeys = new Map<string, string>([
        ['disposed', 'number'],
    ])
    for (const [key, expType] of optionalSimpleKeys) {
        if ((key in input) && typeof input[key] !== expType) {
            throw new Error('Grain Water Jar assertion failure: ' + key + 'was not type ' + expType + '. Was ' + (typeof input[key]));
        }
    }
    // complex optional array keys
    const complexOptionalArrayKeys = new Map<string, (v: any) => boolean>([
        ['notes', IsValidNote],
    ])
    for (const [key, validator] of complexOptionalArrayKeys) {
        if (!OptionalArrayOfType(key, input, validator)) {
            throw new Error('Grain Batch assertion failure: optional array key ' + key + ' was not valid');
        }
    }
    // Unmarshal ACL
    if (!('acl' in input)) {
        throw 'ACL missing from input in asserter'
    }
    input.acl = UnmarshalAcl(input.acl)
    return
}

export default function GrainWaterJarDisplay( // TODO; this whole thing!
    {
        readonly, data, headerLevel, isTopLevel
    }: DisplayInput<GrainWaterJarData>) {
        const {dispatch} = useModalContext();
        const [initial, setInitial] = useState(data)

        const [err, setErr] = useState<string | undefined>()

        // grain batch non-changeable (base grain)
        const [notes, setNotes] = useState<AllEntries<Note>>(InitialNotesState(initial.notes))
        const [disposed, setDisposed] = useState(initial.disposed)
        const [acl, setAcl] = useState<ACL>(initial.acl)
        const updateInitial = (updated: GrainWaterJarData) => {
            setInitial(updated)
            // TODO: more here
            setDisposed(updated.disposed)
            setNotes(InitialNotesState(updated.notes))
            setAcl(updated.acl)
            setErr(undefined)
        }
        const cookies = useContext(CookiesContext)
        const submit = () => {
            const body: any = {
                disposed: disposed,
                notes: notes,
                acl: MarshalAcl(acl),
            }
            DoUpdateRequest("grainWaterJar", initial._id, body, AssertGrainWaterJar, allCookies(cookies))
                .then(v=>{
                    updateInitial(new GrainWaterJarData(v))
                    dispatch({type: ActionTypes.SET_MODAL_INFO, payload:{
                            header: "Update Success",
                            text: "entry updated successfully",
                            isErr: false
                        }})
                })
                .catch(e=>{
                    setErr("failed to update initial: "+JSON.stringify(e))
                    dispatch({type: ActionTypes.SET_MODAL_INFO, payload:{
                            header: "Update Failed",
                            text: "failed to update: " + JSON.stringify(e),
                            isErr: true
                        }})
                })
        }
        const ovcs: OnViewCreatorQuadCol[] = [
            // { // TODO: anything here?
            //     txt: "Create Jars From Batch",
            //     // TODO: creates either PC-d or un-pc'd jars!
            //     // TODO: does this creation need a pcRun??? Can we do it before the run?
            //     // TODO: can items be added when creating a PC run?
            //     newCreationArea: (onCreate: AddCreatedTriColFunction) => {
            //         return <NewJarForm grainBatchIn={initial} handlers={{
            //             onCreate: (newItem: JarData) => {
            //                 return onCreate([{
            //                     typeText: "Grain Jar",
            //                     node: <CreatedLinkFor linkId={newItem._id} typ={"jar"}/>
            //                 }], false)
            //             },
            //             isTopLevel: false,
            //         }}/>
            //     },
            // }
        ]
        return <DisplayFormWrapper entryType={"grainWaterJar"}>
            <ErrorDisplay err={err}/>
            <ID props={{id:data._id, txt:"Grain Water Jar", entryType:"grainWaterJar", linkPage:false, allowOpenMainPage:false}}/>
            <OnViewCreatorsTriColArea OnViewCreators={ovcs} readonly={readonly}/>
            <FlexedArea>
                <FlexedSinglesGroup>
                    <GrainBatchArea batchId={data.grainBatch}/>
                    <DisposedDisplay readonly={readonly} initial={initial.disposed} setDisposedOnParent={setDisposed}/>
                </FlexedSinglesGroup>
                <FlexedSinglesGroup>
                    <DateArea pre={"Created: "} when={initial.creationDate} readonly={true}/>
                    <DateArea pre={"Last Updated: "} when={initial.lastUpdated} readonly={true}/>
                </FlexedSinglesGroup>
            </FlexedArea>
            <NotesFormArea readonly={readonly} initial={initial.notes} updateParent={setNotes}/>
            <TogglableAreaWithDepth startOpen={false} openTxt={"view permissions"} closeTxt={"minimize perms area"}>
                <AclDisplay initial={initial.acl} readonly={readonly} updateParent={setAcl}/>
            </TogglableAreaWithDepth>
            {readonly ? null : <button className={"bottomButton greenButton"} onClick={(e) => {
                e.stopPropagation();
                submit()
            }}>{"Update"}</button>}
        </DisplayFormWrapper>
}

export function NewGrainWaterJarForm({handlers, recipe}: {
    handlers: NewEntryInput<GrainWaterJarData>,
    recipe?: JarRecipeData
}) {
    const {dispatch} = useModalContext();
    const [jarRecipe, setJarRecipe] = useState(recipe)
    const [notes, setNotes] = useState<Note[]>([])
    const [acl, setAcl] = useState<ACL>({blanketPerm:true})
    const [err, setErr] = useState<string | undefined>()

    const baseAcl = NewAllCanWriteAcl()

    const cookies = useContext(CookiesContext)
    const newGrainBatchSubmit = () => {
        if (jarRecipe === undefined) {
            setErr("jarRecipe must exist")
            return
        }
        const body: any = {
            recipe: jarRecipe?._id,
            notes: notes,
            acl: MarshalAcl(acl),
        }
        DoCreateRequest("grainWaterJar", body, AssertGrainWaterJar, allCookies(cookies))
            .then(v=>{
                handlers.onCreate ? handlers.onCreate(new GrainWaterJarData(v)) : console.log("no onCreate provided")
                dispatch({type: ActionTypes.SET_MODAL_INFO, payload:{
                        header: "Create Success",
                        text: "entry created successfully",
                        isErr: false
                    }})
            })
            .catch(e=>{
                setErr(JSON.stringify(e))
                dispatch({type: ActionTypes.SET_MODAL_INFO, payload:{
                        header: "Create Failure",
                        text: "entry failed to create: " + JSON.stringify(e),
                        isErr: true
                    }})
            })
    }
    return <NewEntryFormWrapper entryType={"grainWaterJar"} isTopLevel={handlers.isTopLevel}>
        <ErrorDisplay err={err}/>
        {recipe === undefined &&
            <JarRecipeSelector doSelect={setJarRecipe} allowCreate={handlers.isTopLevel}
                               creatorInPage={handlers.isTopLevel}/>}
        <NewEntryNotes setNotes={setNotes}/>
        <AclDisplay readonly={false} updateParent={setAcl} initial={baseAcl} />
        <button className={"bottomButton greenButton"} onClick={(e) => {
            e.stopPropagation();
            newGrainBatchSubmit()
        }}>{"Update"}</button>
    </NewEntryFormWrapper>
}

export function GrainWaterJarListPageTable({data, onClick, withLink}: ListPageItems<GrainWaterJarData>) {
    let cols: ListTableColumn<GrainWaterJarData>[] = [
        NewColumn("ID", (v) => v._id, true),
        NewColumn("Created", (v) => {
            return NumberToDateStr(v.creationDate)
        }, true),
        NewColumn("Updated", (v) => {
            return NumberToDateStr(v.lastUpdated)
        }),
        // TODO: add more! batch and stuff
    ]
    if (withLink) {
        cols = [...cols, NewColumn("Link", (v: GrainWaterJarData) => {
            return <EntryLinkWrapper props={{entry:v, openInNewTab: true}}>
                <button className={"basicButtonSmall"}>{"View"}</button>
            </EntryLinkWrapper>
        })]
    }
    return <ListPageTable cols={cols} data={data} onClick={onClick} newClass={v=>{return new GrainWaterJarData(v)}}/>
}

export function GrainWaterJarSelectorTable({data, onClick}: ListPageItems<GrainWaterJarData>) {
    return <GrainWaterJarListPageTable data={data} onClick={onClick} withLink={true}/>
}

export function GrainWaterJarSelector(
    {
        doSelect,
        allowCreate
    }: {
        doSelect: (val: GrainWaterJarData | undefined) => void,
        allowCreate?: boolean,
    }) {
    const table = (items: GrainWaterJarData[]): JSX.Element => {
        return <GrainWaterJarSelectorTable data={items} onClick={doSelect}/>
    }

    return <ExistingRecentSelector entryType={"grainWaterJar"} entryTypes={"grainWaterJars"} doSelect={doSelect}
                                   asserter={AssertGrainWaterJar}
                                   table={table}>
        {allowCreate && <NewGrainWaterJarForm handlers={{onCreate: doSelect, isTopLevel: false}}/>}
    </ExistingRecentSelector>
}