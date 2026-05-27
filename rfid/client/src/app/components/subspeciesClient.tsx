'use client'

import React, {JSX, useEffect, useState} from "react";
import {IsValidNote, NewEntryNotes, Note, NotesAreaInline} from "@/app/components/formSubcomponents/notes";
import {AllEntries, Data} from "@/app/components/formSubcomponents/shared";
import ID from "@/app/components/formSubcomponents/id";
import DateArea from "@/app/components/formSubcomponents/date";
import {SubspeciesData} from "@/app/components/subspeciesServer";
import {
    AssertArrayResult,
    CreateNewEntryButton,
    DisplayInput,
    HandleJsonResponse,
    InlineExpansionArea, InlineExpansionButton,
    InlineProps,
    InlineSubArea,
    IsString, ListPageItems,
    NewEntryInput,
    OptionalArrayOfType,
    OptionalKey,
} from "@/app/components/common";
import {AliasesArea, ErrorDisplay, NameArea, SpeciesArea} from "@/app/components/formSubcomponents/commonClient";
import {BaseExternalUrl} from "@/app/components/Constants";
import {ExistingSpeciesSelector} from "@/app/components/speciesClient";
import {
    AclDefaultAclDisplay,
    AclDisplay,
    DefaultAclDisplay,
    IsValidAcl,
    TogglableAreaWithDepth
} from "@/app/components/accessControlClient";
import {ACL} from "@/app/components/accessControlServer";
import TestAndValidate from "@/app/components/testing/untested";
import {HandleErr} from "@/app/components/userClient";
import {SpeciesData} from "@/app/components/speciesServer";
import {DisplayFormWrapper, NewEntryFormWrapper, Subform} from "@/app/components/lcRecipeClient";
import {DepthProvider} from "@/app/components/formSubcomponents/depthContext/depth";
import {ExistingRecentSelector, InlineEntry} from "./agarRecipeClient";
import {
    FlexedArea,
    FlexedSinglesGroup, ListPageTable,
    ListTableColumn,
    NewColumn,
    NotesFormArea, NumberToDateStr
} from "@/app/components/agarBatchClient";
import {InitialNotesState} from "@/app/components/formSubcomponents/contaminations";
import {StasisTubeData} from "@/app/components/stasisTubeServer";
import {FruitingChamberData} from "@/app/components/fruitingChamberServer";
import {EntryLinkWrapper} from "@/app/components/formSubcomponents/entryLink";
import {SlantData} from "@/app/components/slantServer";
import {AssertSlant, NewSlantForm} from "@/app/components/slantClient";
import {AssertSubRecipeListResult} from "@/app/components/substrateRecipeClient";

export function AssertSubspecies(input: any): asserts input is SubspeciesData {
    if (typeof input !== 'object') {
        throw new Error('Input is not an object! Input is ' + typeof input);
    }
    // required simple keys
    let requiredSimpleKeys = new Map<string, string>([
        ['_id', 'string'],
        ['species', 'string'],
        ['lastUpdated', 'number'],
    ])
    for (let [key, expType] of requiredSimpleKeys) {
        if (!(key in input && typeof input[key] === expType)) {
            throw new Error('Subspecies assertion failure: ' + key + 'was not type ' + expType + '. Was ' + (typeof input[key]));
        }
    }
    // complex optional keys
    let complexOptionalKeys = new Map<string, (v: any) => boolean>([
        ['acl', IsValidAcl],
        ['defaultAcl', IsValidAcl]
    ])
    for (let [key, validator] of complexOptionalKeys) {
        if (!OptionalKey(key, input, validator)) {
            throw new Error('Subspecies assertion failure: optional key ' + key + ' was not valid: ');
        }
    }
    // complex optional array keys
    let complexOptionalArrayKeys = new Map<string, (v: any) => boolean>([
        ['notes', IsValidNote],
        ['aliases', IsString],
    ])
    for (let [key, validator] of complexOptionalArrayKeys) {
        if (!OptionalArrayOfType(key, input, validator)) {
            throw new Error('Subspecies assertion failure: optional array key ' + key + ' was not valid');
        }
    }
    return
}

export default function SubspeciesDisplay(
    {
        id, readonly, data, headerLevel, isTopLevel, cookies
    }: DisplayInput) {
    try {
        AssertSubspecies(data) // TODO: ENSURE ACL IS BEING PARSED CORRECTLY
        const [initial, setInitial] = useState(data)

        const [aliases, setAliases] = useState(data.aliases || [])
        const [notes, setNotes] = useState<AllEntries<Note>>(InitialNotesState(initial.notes))
        const [err, setErr] = useState<string | undefined>()
        const [acl, setAcl] = useState<ACL | undefined>(initial.acl)
        const [defaultAcl, setDefaultAcl] = useState<ACL | undefined>(initial.defaultAcl)
        const updateInitial = (updated: SubspeciesData) => {
            setAliases(updated.aliases || [])
            setNotes(InitialNotesState(updated.notes))
            setAcl(updated.acl)
            setDefaultAcl(updated.defaultAcl)
        }
        const update = () => {
            fetch(BaseExternalUrl + "/db/update/subspecies/"+encodeURI(initial._id), {
                method: "POST",
                headers: {
                    credentials: 'include',
                    'Content-type': 'application/json'
                },
                body: JSON.stringify({
                    aliases: aliases,
                    notes: notes,
                    acl: acl,
                    defaultAcl: defaultAcl,
                })
            })
                .then(HandleJsonResponse)
                .then((entry) => {
                    AssertSubspecies(entry)
                    updateInitial(entry)
                })
                .catch((error) => {
                    HandleErr(error, setErr)
                });
        }
        return (
            <DisplayFormWrapper entryType={"subspecies"}>
                <ErrorDisplay err={err} headerLevel={headerLevel}/>
                <TestAndValidate todos={["Species up here too?"]}>
                    <ID id={data._id} txt={"Subspecies"} entryType={"subspecies"} />
                </TestAndValidate>
                <FlexedArea>
                    <FlexedSinglesGroup>
                        <DateArea pre={"Last Updated: "} when={initial.lastUpdated} readonly={true}/>
                    </FlexedSinglesGroup>
                </FlexedArea>
                <AliasesArea aliases={aliases} readonly={readonly} headerLevel={headerLevel} updateParent={setAliases}/>
                <NotesFormArea readonly={readonly} initial={initial.notes} updateParent={setNotes}/>
                <AclDefaultAclDisplay ACL={acl} defaultACL={defaultAcl} updateAcl={setAcl} updateDefaultAcl={setDefaultAcl} readonly={readonly}/>
                {readonly ? null : <button className={"bottomButton greenButton"} onClick={(e)=>{
                    e.stopPropagation();
                    update()
                }}>{"Update Subspecies"}</button>}
            </DisplayFormWrapper>
        )
    } catch (err) {
        return <div>{"ERROR: Subspecies data format incorrect: " + err}</div>
    }

}

export function NewSubspeciesForm({handlers, species}: {
    handlers: NewEntryInput<SubspeciesData>,
    species?: SpeciesData
}) {
    const {onCreate} = handlers
    const [name, setName] = useState<string | undefined>()
    const [selectedSpecies, setSelectedSpecies] = useState(species)
    const [aliases, setAliases] = useState<string[]>([])
    const [notes, setNotes] = useState<Note[]>([])
    const [err, setErr] = useState<string | undefined>(undefined)
    const submitNewSubspecies = () => {
        if (!name) {
            setErr("Name must note be blank")
            return
        }
        if (!selectedSpecies) {
            setErr("Species must be selected")
            return
        }
        fetch(BaseExternalUrl + "/create/subspecies", {
            method: "POST",
            headers: {
                credentials: 'include',
                'Content-type': 'application/json'
            },
            body: JSON.stringify({
                name: name,
                species: selectedSpecies,
                aliases: aliases,
                notes: notes,
            })
        })
            .then(HandleJsonResponse)
            .then((entry) => {
                AssertSubspecies(entry)
                onCreate && onCreate(entry)
            })
            .catch((error) => {
                setErr(JSON.stringify(error))
            });
    }
    return (
        <NewEntryFormWrapper entryType={"subspecies"}>
            <ErrorDisplay err={err}/>
            {species === undefined && <ExistingSpeciesSelector initialSpecies={species} doSelect={s => {
                setSelectedSpecies(s)
            }} />}
            {/* NAME (ID) */}
            <NameArea classNames={"inlineChildren"} currentName={name} headerTxt={"New Subspecies Name: "} setName={setName} readonly={false}/>
            {/* Aliases */}
            <AliasesArea aliases={aliases} readonly={false} updateParent={setAliases}/>
            {/* Notes */}
            <NewEntryNotes setNotes={setNotes}/>
            <CreateNewEntryButton onSubmit={submitNewSubspecies}/>
        </NewEntryFormWrapper>
    )
}

// export function SubspeciesInline(
//     {props, showSpeciesName}: {
//         props: InlineProps<SubspeciesData>,
//         showSpeciesName: boolean,
//     }) { // TODO: TEST FOR SHOWSPECIESNAME==true!
//     const aliases = props.data.aliases || []
//     const notes = props.data.notes || []
//     const [expanded, setExpanded] = useState(props.expandByDefault)
//     return <InlineEntry onClick={props.onClick}><TestAndValidate todos={["ensure working and looks good"]}>
//         <InlineSubArea props={{}}>
//             <ID id={props.data._id} txt={"Subspecies"} entryType={"subspecies"} allowOpenMainPage={props.showMainPageButton} linkPage={props.idIsLink}/>
//             {showSpeciesName &&
//                 <SpeciesArea readonly={true} initial={props.data.species} headerLevel={props.headerLevel}/>}
//             {/* Aliases */}
//             <div className={"ml-[2em]"}>
//                 {aliases.map((alias, i) => {
//                     return <div key={i}>{alias}</div>
//                 })}
//             </div>
//         </InlineSubArea>
//         <InlineExpansionArea props={{expanded: expanded}}>
//             {/* Notes */}
//             <NotesAreaInline notes={notes} header={"Notes"} offset={-1}/>
//             {/* Last Updated */}
//             <DateArea pre={"Last Updated: "} when={props.data.lastUpdated} readonly={true}/>
//         </InlineExpansionArea><InlineExpansionButton data-cy-id="InlineSubAreaButton" setExpanded={setExpanded}
//                                expanded={expanded}/>
//     </TestAndValidate>
//     </InlineEntry>
// }

// ExistingSubSpeciesSelector selects between subspecies of a SINGLE species!
export function ExistingSubSpeciesSelector(
    {
        species,
        doSelect,
        headerLevel,
    }: {
        species?: string,
        doSelect: (val: SubspeciesData | undefined) => void,
        headerLevel?: number
    }) {
    const [isLoaded, setLoaded] = useState(false)
    const [selectable, setSelectable] = useState(false)
    const [selectorOpen, setSelectorOpen] = useState(false)
    const [subspeciesList, setSubspeciesList] = useState<SubspeciesData[]>([]);
    const [selected, setSelected] = useState<SubspeciesData | undefined>(undefined)
    const [err, setErr] = useState<string | undefined>(undefined)
    useEffect(() => {
        if (!species){
            return
        }
        setSelected(undefined)
        setLoaded(false)
        fetch(BaseExternalUrl + "/subspeciesFor/" + encodeURI(species), {
            method: "GET",
            headers: {
                credentials: 'include',
                'Accept': 'application/json',
            },
        })
            .then(HandleJsonResponse)
            .then((data) => {
                AssertArrayResult(data, AssertSubspecies)
                setSubspeciesList(data as SubspeciesData[])
                setLoaded(true)
                setSelectable(species !== undefined)
                setErr(undefined)
            })
            .catch((error) => {
                setErr(JSON.stringify(error))
            });
    }, [species]);
    let errArea = () => {
        return <ErrorDisplay err={err} headerLevel={headerLevel}/>
    }
    const toggleButton = () => {
        return <div>
            <button className={"basicButton"} onClick={() => {
                setSelectorOpen(!selectorOpen)
            }}>{selectorOpen ? "Close subspecies selector" : (selected ? "Choose a different subspecies" : "Select a subspecies")}</button>
        </div>
    }
    if (!selectable) {
        return null
    }
    if (!selected && !selectorOpen) {
        return <div className={"centerHChildren gapTop gapBottom"}>
            {errArea()}
            {toggleButton()}
        </div>
    }
    if (!isLoaded) {
        return <div className={"centerHChildren gapTop gapBottom"}>
            <ErrorDisplay err={err} headerLevel={headerLevel}/>
            <div>{"loading subspecies selector"}</div>
        </div>
    }
    if (subspeciesList.length == 0) {
        return <div className={"centerHChildren gapTop gapBottom"}>
            <ErrorDisplay err={"No Subspecies Found for species: " + (species && species)} headerLevel={headerLevel}/>
            <TestAndValidate todos={["do this"]}>
                <div>{"CREATE SUBSPECIES LINK"}</div>
            </TestAndValidate>
        </div>
    }
    if (selected && !selectorOpen) {
        return <div className={"centerHChildren gapTop gapBottom"}>
            {errArea()}
            {"Currently Selected subspecies: " + selected._id}
            {toggleButton()}
        </div>
    }

    return <div className={"centerHChildren gapTop gapBottom"}>
        <Subform>
        {errArea()}
        {toggleButton()}
        <SubspeciesSelector doSelect={s=>{
            setSelected(s)
            setSelectorOpen(false)
            doSelect(s)
        }} />
        {toggleButton()}
        </Subform>
    </div>
}

export function SubspeciesFormArea({subspecies}:{
    subspecies: string,
}){ // TODO: validate link works
    return <EntryLinkWrapper props={{entryType:"subspecies",linkId:encodeURI(subspecies)}}><div>{"Subspecies: "+subspecies}</div></EntryLinkWrapper>
}

export function SubspeciesListPageTable({data, onClick, withLink}: ListPageItems<SubspeciesData>) {
    let cols: ListTableColumn<SubspeciesData>[] = [
        NewColumn("Subspecies", (v)=>v._id),
        NewColumn("Species", (v)=>v.species),
        NewColumn("Updated", (v)=>{
            return NumberToDateStr(v.lastUpdated)
        }),
    ]
    if (withLink) {
        cols = [...cols, NewColumn("Link", (v: SubspeciesData)=>{
            return <EntryLinkWrapper props={{linkId:encodeURI(v._id),entryType:"subspecies",openInNewTab:true}}>{/* TODO: ensure ok*/}
                <button className={"basicButtonSmall"}>{"View"}</button>
            </EntryLinkWrapper>
        })]
    }
    return <ListPageTable cols={cols} data={data} onClick={onClick}/>
}
export function SubspeciesSelectorTable({data, onClick}: ListPageItems<SubspeciesData>) {
    return <SubspeciesListPageTable data={data} onClick={onClick} withLink={true} />
}

// SubspeciesSelector is a selector between ALL subspecies, not just those of a single species
export function SubspeciesSelector( // TODO: USE ELSEWHERE
    {
        doSelect,
        allowCreate
    }: {
        doSelect: (val: SubspeciesData | undefined) => void,
        allowCreate?: boolean
    }) {
    const table = (items: SubspeciesData[]):JSX.Element=>{
        return <SubspeciesSelectorTable data={items} onClick={doSelect}/>
    }

    return <ExistingRecentSelector entryType={"subspecies"} entryTypes={"subspecies"} doSelect={doSelect} asserter={AssertSubspecies}
                                   table={table}>
        {/* TODO: ok? allowCreate && <NewSubspeciesForm handlers={{onCreate: doSelect,isTopLevel: false}}/>*/}
    </ExistingRecentSelector>
}