'use client'

import React, {JSX, useEffect, useState} from "react";
import {IsValidNote, NewEntryNotes, Note, NotesAreaInline} from "@/app/components/formSubcomponents/notes";
import {AllEntries, Data} from "@/app/components/formSubcomponents/shared";
import ID from "@/app/components/formSubcomponents/id";
import DateArea from "@/app/components/formSubcomponents/date";
import {SubspeciesData} from "@/app/components/subspeciesServer";
import {
    CreateNewEntryButton,
    DisplayInput,
    HandleJsonResponse,
    InlineExpansionArea, InlineExpansionButton,
    InlineProps,
    InlineSubArea,
    IsString,
    NewEntryInput,
    OptionalArrayOfType,
    OptionalKey,
} from "@/app/components/common";
import {AliasesArea, ErrorDisplay, NameArea, SpeciesArea} from "@/app/components/formSubcomponents/commonClient";
import {BaseExternalUrl} from "@/app/components/Constants";
import {ExistingSpeciesSelector} from "@/app/components/speciesClient";
import {
    AclDisplay,
    DefaultAclDisplay,
    IsValidAcl,
    TogglableAreaWithDepth
} from "@/app/components/accessControlClient";
import {ACL} from "@/app/components/accessControlServer";
import TestAndValidate from "@/app/components/testing/untested";
import {HandleErr} from "@/app/components/userClient";
import {SpeciesData} from "@/app/components/speciesServer";
import {DisplayFormWrapper, NewEntryFormWrapper} from "@/app/components/lcRecipeClient";
import {DepthProvider} from "@/app/components/formSubcomponents/depthContext/depth";
import { InlineEntry } from "./agarRecipeClient";
import {FlexedArea, FlexedSinglesGroup, NotesFormArea} from "@/app/components/agarBatchClient";
import {InitialNotesState} from "@/app/components/formSubcomponents/contaminations";
// TODO: list page not working
// TODO: ensure display page doing what we want

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
                {/* TODO: ACL/DefaultACL buttons side-by-side*/}
                <TestAndValidate todos={["make these side-by-side?"]}>
                    <TogglableAreaWithDepth startOpen={false} openTxt={"view permissions"}
                                   closeTxt={"minimize perms area"}>
                        <AclDisplay ACL={acl} readonly={readonly} updateParent={setAcl}/>
                    </TogglableAreaWithDepth>
                    <TogglableAreaWithDepth startOpen={false} openTxt={"view default ACL"}
                                   closeTxt={"minimize default ACL area"}>
                        <DefaultAclDisplay readonly={readonly} ACL={defaultAcl} updateParent={setDefaultAcl}/>
                    </TogglableAreaWithDepth>
                </TestAndValidate>
                {readonly ? null : <button className={"bottomButton greenButton"} onClick={(e)=>{
                    e.stopPropagation();
                    update()
                }}>{"Update"}</button>}
                {/* TODO: unlikely: <OnViewCreatorsQuadColArea OnViewCreators={ovcs} readonly={readonly}/>*/}
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
    // TODO: handle isTopLevel
    const submitNewSubspecies = () => {
        if (!name) {
            setErr("Name must note be blank")
            return
        }
        if (!selectedSpecies) {
            setErr("Species must be selected")
            return
        }
        fetch(BaseExternalUrl + "/create/subspecies", { // TODO: ensure correct
            method: "POST",
            headers: {
                credentials: 'include',
                // TODO: may need 'Cookie': cookies,
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
                // TODO: anything else?
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

export function SubspeciesInline(
    {props, showSpeciesName}: {
        props: InlineProps<SubspeciesData>,
        showSpeciesName: boolean,
    }) { // TODO: TEST FOR SHOWSPECIESNAME==true!
    const aliases = props.data.aliases || []
    const notes = props.data.notes || []
    const [expanded, setExpanded] = useState(props.expandByDefault)
    return <InlineEntry onClick={props.onClick}><TestAndValidate todos={["ensure working and looks good"]}>
        <InlineSubArea props={{}}>
            <ID id={props.data._id} txt={"Subspecies"} entryType={"subspecies"} allowOpenMainPage={props.showMainPageButton} linkPage={props.idIsLink}/>
            {showSpeciesName &&
                <SpeciesArea readonly={true} initial={props.data.species} headerLevel={props.headerLevel}/>}
            {/* Aliases */}
            <div className={"mleft"}>{/* TODO: CHANGE@! */}
                {aliases.map((alias, i) => {
                    // TODO: HIDE SOME ALIASES IF TOO MANY?
                    return <div key={i}>
                        <div>{alias}</div>
                    </div>
                })}
            </div>
        </InlineSubArea>
        <InlineExpansionArea props={{expanded: expanded}}>
            {/* Notes */}
            <NotesAreaInline notes={notes} header={"Notes"} offset={-1}/>
            {/* Last Updated */}
            <DateArea pre={"Last Updated: "} when={props.data.lastUpdated} readonly={true}/>
        </InlineExpansionArea><InlineExpansionButton data-cy-id="InlineSubAreaButton" setExpanded={setExpanded}
                               expanded={expanded}/>
    </TestAndValidate>
    </InlineEntry>
}

export function ExistingSubSpeciesSelector( // TODO: ensure perms are followed
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
    const subspeciesFor: (s: string) => SubspeciesData = (s: string) => { // TODO: DELETEME
        return {
            _id: s,
            species: "SPECIES_NAME",
            aliases: ["alias1", "alias2", "alias3"],
            notes: [{time: 0, note: "NOTE 1"}, {time: 0, note: "NOTE 2"}],
            lastUpdated: 0,
            //perms: {userPerms: {ids:[],canWrite:[]}, projectPerms: {ids:[],canWrite:[]}, blanketPerms: 2} // TODO: OK?
        }
    }
    useEffect(() => {
        //setSelected(undefined)
        // setSubspeciesList([subspeciesFor('subs_A'), subspeciesFor('subs_B')]) // TODO: REMOVE
        // setLoaded(true) // TODO: REMOVE
        // setSelectable(species !== undefined) // TODO: REMOVE
        // return // TODO: REMOVE
        // if (species === undefined) {
        //     setSelectable(false)
        //     return
        // }
        fetch(BaseExternalUrl + "/db/list/subspeciesFor/" + species, { // TODO: ensure correct
            method: "GET",
            headers: {
                credentials: 'include', // TODO: check that user has creds for species
                //'Cookie': cookies,
                // TODO: THIS!
            },
        })
            .then(HandleJsonResponse)
            .then((data) => {
                setSubspeciesList(data as SubspeciesData[]) // TODO: ASSERT????
                setLoaded(true)
                // setSelectable(species !== undefined)
                setErr(undefined)
            })
            .catch((error) => {
                setErr(JSON.stringify(error))
            });
    }, [species]);
    // TODO: CLEAR SELECTION
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
        return null;
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
            {/* TODO: CREATE SUBSPECIES BUTTON*/}
            <div>{"CREATE SUBSPECIES LINK"}</div>
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
        {errArea()}
        {toggleButton()}
        <div className={"fullWidth"}>
            <DepthProvider>{/* TODO: necessary?*/}
            {subspeciesList.map((sub, i) => {
                // TODO: MAKE SURE THIS IS WIDE ENOUGH!
                return <SubspeciesInline key={i} showSpeciesName={species === undefined} props={{
                    data: sub, onClick: (subsp) => {
                        doSelect(subsp)
                        setSelectorOpen(false)
                        setSelected(subsp)
                    }
                }
                }/>
            })}
            </DepthProvider>
        </div>
        {toggleButton()}
    </div>
}

export function SubspeciesFormArea({subspecies}:{
    subspecies: string,
}){ // TODO: LINK!
    return <div>{"Subspecies: "+subspecies}</div>
}