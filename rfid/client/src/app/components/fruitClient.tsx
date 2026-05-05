'use client'

import React, {useState} from "react";
import {IsValidNote, NewEntryNotes, Note, NotesAreaInline} from "@/app/components/formSubcomponents/notes";
import {
    AddCreatedQuadColFunction,
    AllEntries,
    OnViewCreatorQuadCol,
    SplitAllEntries
} from "@/app/components/formSubcomponents/shared";
import ID from "@/app/components/formSubcomponents/id";
import DateArea from "@/app/components/formSubcomponents/date";
import {
    InitialPicsEntries,
    IsValidPicWithNotesIncoming,
    NewPicWithNotesForm,
    PicWithNotesForm,
} from "@/app/components/formSubcomponents/picWithNotes";
import {AddToTransfers, TransfersOutDisplay} from "@/app/components/transferClient";
import {
    HandleJsonResponse,
    HandleTxtResponse,
    ImportDisplayInput,
    InlineExpansionArea,
    InlineExpansionButton,
    InlineProps,
    InlineSubArea,
    IsString, ListPageItems,
    OptionalArrayOfType,
    OptionalKey,
    OptionalSimpleKey,
    resolvePicsFormData,
    SendMultipartRequest,
    setFormData,
    setFormImages
} from "@/app/components/common";
import {
    DisposedDisplay,
    ErrorDisplay,
    GensFormDisplay,
    GensInlineDisplay,
    MostRecentImageDisplay,
    NameArea,
    ParentDisplay,
    PicsDisplay,
    SpeciesArea,
    SubspeciesArea
} from "@/app/components/formSubcomponents/commonClient";
import {FruitData} from "@/app/components/fruitServer";
import ImageSelector from "@/app/components/formSubcomponents/imageSelector";
import {redirect} from "next/navigation";
import EntryLink from "@/app/components/formSubcomponents/entryLink";
import {NewSporePrintForm} from "@/app/components/sporePrintClient";
import {BaseExternalUrl} from "@/app/components/Constants";
import {SpeciesData} from "@/app/components/speciesServer";
import {SubspeciesData} from "@/app/components/subspeciesServer";
import {ReadRFIDButton} from "@/app/components/formSubcomponents/readerWriterButtons/readerSelector";
import {ExistingSpeciesSelector} from "@/app/components/speciesClient";
import {ExistingSubSpeciesSelector} from "@/app/components/subspeciesClient";
import TestAndValidate from "@/app/components/testing/untested";
import {AclDisplay, IsValidAcl, TogglableAreaWithDepth} from "@/app/components/accessControlClient";
import {ACL} from "@/app/components/accessControlServer";
import {OvcForXfers} from "@/app/components/bagClient";
import {OnViewCreatorsQuadColArea} from "@/app/components/pcRunClient";
import {NewSporeSwabForm} from "@/app/components/sporeSwabClient";
import {SporeSwab} from "@/app/components/sporeSwabServer";
import {CreatedLinkFor} from "@/app/components/substrateRecipeClient";
import {RecentSelectorV2} from "@/app/components/mssClient";
import {SporePrintData} from "@/app/components/sporePrintServer";
import {DisplayFormWrapper, ImportEntryFormWrapper, NewEntryFormWrapper} from "@/app/components/lcRecipeClient";
import {InlineEntry} from "@/app/components/agarRecipeClient";
import {
    FlexedArea,
    FlexedSinglesGroup, ListPageTable,
    ListTableColumn,
    NewColumn,
    NotesFormArea, NumberToDateStr
} from "@/app/components/agarBatchClient";
import {CreatedUpdatedDisposedArea} from "@/app/components/plateClient";
import {InitialNotesState} from "@/app/components/formSubcomponents/contaminations";
import {SpeciesSubspeciesArea} from "@/app/components/lcClient";
import {BagData} from "@/app/components/bagServer";

export function AssertFruit(input: any): asserts input is FruitData {
    if (typeof input !== 'object') {
        throw new Error('Input is not an object! Input is ' + typeof input);
    }

    // required simple keys
    let requiredSimpleKeys = new Map<string, string>([
        ['_id', 'string'],
        ['creationDate', 'number'],
        ['species', 'string'],
        ['lastUpdated', 'number'],
    ])
    for (let [key, expType] of requiredSimpleKeys) {
        if (!(key in input && typeof input[key] === expType)) {
            throw new Error('Bag assertion failure: ' + key + 'was not type ' + expType + '. Was ' + (typeof input[key]));
        }
    }

    // optional simple keys
    let optionalSimpleKeys = new Map<string, string>([
        ['subspecies', 'string'],
        ['genSpore', 'number'],
        ['parentType', 'string'],
        ['parent', 'string'],
        ['disposed', 'number'],
    ])
    for (let [key, expType] of optionalSimpleKeys) {
        if (!OptionalSimpleKey(key, input, expType)) {
            throw new Error('Bag assertion failure: optional key ' + key + ' was not valid');
        }
    }

    // complex optional keys
    let complexOptionalKeys = new Map<string, (v: any) => boolean>([
        ['mostRecentImage', IsValidPicWithNotesIncoming],
        ['acl', IsValidAcl]
    ])
    for (let [key, validator] of complexOptionalKeys) {
        if (!OptionalKey(key, input, validator)) {
            throw new Error('Fruit assertion failure: optional key ' + key + ' was not valid');
        }
    }
    // complex optional array keys
    let complexOptionalArrayKeys = new Map<string, (v: any) => boolean>([
        ['transfersOut', IsString],
        ['prints', IsString],
        ['pics', IsValidPicWithNotesIncoming],
        ['notes', IsValidNote],
    ])
    for (let [key, validator] of complexOptionalArrayKeys) {
        if (!OptionalArrayOfType(key, input, validator)) {
            throw new Error('Bag assertion failure: optional array key ' + key + ' was not valid');
        }
    }
    return
}

export default function FruitDisplay(
    {
        id, readonly, data, headerLevel, openSporesInNewTab, allowPrintCreation, isTopLevel, cookies
    }: {
        id: string;
        readonly: boolean;
        isTopLevel: boolean;
        data: any;
        headerLevel?: number;
        openSporesInNewTab?: boolean;
        allowPrintCreation?: boolean;
        cookies: string;
    }) {
    try {
        AssertFruit(data)
        const [initial, setInitial] = useState(data)
        // TODO: change all other states when re-set

        const [pics, setPics] = useState<SplitAllEntries<PicWithNotesForm, NewPicWithNotesForm>>(InitialPicsEntries(initial.pics))
        const [disposed, setDisposed] = useState(initial.disposed)
        const [notes, setNotes] = useState<AllEntries<Note>>(InitialNotesState(initial.notes))
        // Helper states
        const [transfersOut, setTransfersOut] = useState(data.transfersOut || [])
        const [sporePrints, setSporePrints] = useState(data.prints || [])
        const [err, setErr] = useState<string | undefined>()
        const [acl, setAcl] = useState<ACL | undefined>(initial.acl)
        const updateInitial = (updated: FruitData) => {
            setInitial(updated)
            setPics(InitialPicsEntries(updated.pics))
            setDisposed(updated.disposed)
            setNotes(InitialNotesState(updated.notes))
            // Helper states
            setTransfersOut(updated.transfersOut || [])
            setSporePrints(updated.prints || [])
            setAcl(updated.acl)
        }
        const sporePrintsArea = () => {
            return <div>
                <div>{"Spore Prints: "}</div>
                {(sporePrints.length === 0) &&
                    <div>{"None"}</div>}
                {sporePrints.map(spid => {
                    const b58id = spid
                    return <div key={b58id}>
                        <EntryLink props={{
                            displayedId: b58id,
                            linkId: b58id,
                            entryType: "sporePrint",
                            openInNewTab: openSporesInNewTab
                        }}>{spid}</EntryLink>
                    </div>
                })}
            </div>
        }
        const fruitSubmit = () => {
            // disposed, notes, existing pics
            let body = new FormData()
            let dataObj: any = {
                notes: notes,
                disposed: disposed,
                acl: acl,
            }
            try {
                // Pics
                let picsInfo = resolvePicsFormData(pics)
                let newImages = picsInfo.images
                dataObj.images = picsInfo.obj
                // Set data on form
                setFormData(body, dataObj)
                //body.set("data", JSON.stringify(dataObj))
                setFormImages(body, "newPic", newImages)
            } catch (caught: any) {
                setErr(JSON.stringify(caught))
                return
            }

            SendMultipartRequest(BaseExternalUrl + "/db/update/fruit/" + initial._id, cookies, body)
                // fetch(BaseExternalUrl + "/db/update/fruit/" + data._id, {
                //     method: 'Post',
                //     body: body,
                //     headers: {
                //         credentials: 'include',
                //         'Cookie': cookies,
                //         // 'Content-type': "multipart/form-data"
                //         //Authorization: tokenFetch, // TODO: auth?
                //     },
                // })
                .then(HandleJsonResponse)
                .then((newEntry) => {
                    AssertFruit(newEntry)
                    updateInitial(newEntry)
                })
                .catch((er) => {
                    setErr(JSON.stringify(er))
                });
        }
        const ovcs: OnViewCreatorQuadCol[] = [
            // TODO: setTransfersOut on this as needed!
            // TODO: USE THIS!
            OvcForXfers(data._id, "fruit", ["plate", "slant", "jar", "stasisTube"], cookies, AddToTransfers(setTransfersOut, transfersOut), "Clone/Transfer Fruit"), // TODO: ensure list correct// TODO: OVC for clone to plate (transfer)
            {
                txt: "Create Spore Swab",
                newCreationArea: (onCreate: AddCreatedQuadColFunction) => {
                    return <NewSporeSwabForm fruitIn={data} onCreate={(item: SporeSwab) => {
                        onCreate([{
                            typeText: "Spore Swab",
                            node: <CreatedLinkFor linkId={item._id} typ={"sporeSwab"}/>,
                        }])
                    }}/>
                }
            },
            {
                txt: "Create Spore Print",
                newCreationArea: (onCreate: AddCreatedQuadColFunction) => {
                    return <NewSporePrintForm fruitIn={data}
                                              cookies={cookies/* TODO: remove cookies and make like others*/}
                                              onCreate={(item: SporePrintData) => {
                                                  onCreate([{
                                                      typeText: "Spore Print",
                                                      node: <CreatedLinkFor linkId={item._id} typ={"sporePrint"}/>,
                                                  }])
                                              }}/>
                }
            }

        ]
        return (
            <DisplayFormWrapper entryType={"fruit"}>
                <ErrorDisplay err={err}/>
                <ID txt={"Fruit"} id={data._id} entryType={"fruit"}/>
                <OnViewCreatorsQuadColArea OnViewCreators={ovcs} readonly={readonly}/>{/* TODO: where to put?*/}
                <MostRecentImageDisplay data={initial.mostRecentImage} headerLevel={headerLevel}/>
                <FlexedArea>
                    <FlexedSinglesGroup>
                        <SpeciesSubspeciesArea species={initial.species} subspecies={initial.subspecies}/>
                        <ParentDisplay parent={initial.parent} parentType={initial.parentType}
                                       headerLevel={headerLevel}/>
                    </FlexedSinglesGroup>
                    <FlexedSinglesGroup>
                        <CreatedUpdatedDisposedArea created={initial.creationDate} updated={initial.lastUpdated}
                                                    readonly={readonly} disposed={initial.disposed}
                                                    setDisposedOnParent={setDisposed}/>
                    </FlexedSinglesGroup>
                    <FlexedSinglesGroup>
                        <GensFormDisplay gensSinceSpore={initial.genSpore} dontDisplayGensFruitOrSpore={true}
                                         headerLevel={headerLevel}/>
                    </FlexedSinglesGroup>
                </FlexedArea>
                <TransfersOutDisplay thisId={initial._id} thisEntryType={"fruit"} transfersOut={transfersOut}
                                     allowNewTransferCreation={false}
                                     cookies={cookies}/>
                <PicsDisplay pix={pics} updateParent={setPics} readonly={readonly}/>{/* Pics */}
                <NotesFormArea readonly={readonly} initial={initial.notes} updateParent={setNotes}/>
                <TogglableAreaWithDepth startOpen={false} openTxt={"view permissions"} closeTxt={"minimize perms area"}>
                    <AclDisplay ACL={acl} readonly={readonly} updateParent={setAcl} />
                </TogglableAreaWithDepth>
                {readonly ? null : <button className={"bottomButton greenButton"} onClick={(e)=>{
                    e.stopPropagation();
                    fruitSubmit()
                }}>{"Update"}</button>}
            </DisplayFormWrapper>
        )
    } catch (err) {
        return <div>{"ERROR: Fruit data format incorrect: " + err}</div>
    }
}

export function NewFruitForm(
    {parentId, parentType, headerLevel, readonly, onCreate, cookies}: {
        parentId: string,
        parentType: string,
        headerLevel?: number,
        readonly: boolean,
        onCreate: (f: FruitData) => void,
        cookies: string
    }) {
    if (readonly) {
        return null
    }
    const [harvestDate, setHarvestDate] = useState(Date.now())
    const [pics, setPics] = useState<NewPicWithNotesForm[]>([])
    const [notes, setNotes] = useState<Note[]>([])
    const [err, setErr] = useState<string | undefined>()
    //const [perms, setPerms] = useState<EntryPerms | undefined>() // TODO: inherit from parents????
    const newFruitSubmit = () => {
        let body = new FormData()
        let dataObj: any = {
            parentId: parentId,
            parentType: parentType,
            harvestDate: harvestDate,
        }
        // TODO: do we need custom perms here? Probably not, inherit
        if (notes.length > 0) {
            dataObj.notes = notes
        }
        if (pics.length > 0) {
            dataObj.pics = pics.map(p => {
                return {
                    time: p.time, notes: p.notes.new.map(v => {
                        return v.data
                    })
                }
            })
            for (let i = 0; i < pics.length; i++) {
                let imgi = pics[i].img
                if (imgi === undefined) {
                    setErr("new image #" + i + " was not set!")
                    return
                }
                const filePrefix = "newPic" + "-" + i
                body.set(filePrefix, imgi, filePrefix)
            }
        }
        setFormData(body, dataObj)
        //body.set("data", dataObj)
        SendMultipartRequest(BaseExternalUrl + "/create/fruit", cookies, body)
            .then(HandleJsonResponse).then((newEntry) => {
            try {
                AssertFruit(newEntry)
                onCreate(newEntry)
                // TODO: WE HAVE CREATED A NEW FRUIT!!!!! NOTIFY???
                // TODO ? redirect(BaseUrl+"/view/fruit/"+newId) // TODO: ok?
            } catch (er) {
                setErr("failed to decode response:")
            }
        }).catch((er) => {
            setErr(JSON.stringify(er))
        });
    }
    return (
        <NewEntryFormWrapper entryType={"fruit"}>
            {/* TODO: TITLE? */}
            <ErrorDisplay err={err} headerLevel={headerLevel}/>
            <DateArea pre={"Harvest Date: "} readonly={false} updateParent={setHarvestDate}/>
            <PicsDisplay updateParent={v => {
                setPics(v.new.map(x => {
                    return x.data
                }))
            }} headerLevel={headerLevel} readonly={false}/>
            <NewEntryNotes setNotes={setNotes}/>

            {/*<EntryPermsArea setEntryPerms={setPerms}/> /!* TODO: perms from parent? *!/*/}
            <input type="submit" value="Submit" onClick={newFruitSubmit} onSubmit={(e) => {
                e.preventDefault();
            }}/>
        </NewEntryFormWrapper>
    )
}

export function FruitImportDisplay({headerLevel, cookies}: ImportDisplayInput) { // TODO: USE ONLY FOR FRUITS PURCHASED OR FOUND
    const [parentType, setParentType] = useState<string | undefined>(undefined) // TODO: ensure this is everywhere in ts and go
    const [species, setSpecies] = useState<SpeciesData | undefined>(undefined)
    const [subspecies, setSubspecies] = useState<SubspeciesData | undefined>(undefined)
    const [imageFile, setImageFile] = useState<File | undefined>(undefined)
    const [notes, setNotes] = useState<Note[]>([])
    const [err, setErr] = useState<string | undefined>(undefined)
    const submitImportFruit = () => { // TODO: rework so we only have the one image, and the one data set
        if (parentType === undefined) {
            setErr("source area must be set!")
            return
        }
        if (parentType !== "store" && parentType !== "outside") { // TODO: ENSURE OK ELSEWHERE
            setErr("parentType must be store or outside!")
            return
        }
        if (species === undefined) {
            setErr("Species must be set!")
            return
        }
        let formData = new FormData()
        let dataObj: any = {
            parentType: parentType,
            species: species._id,
            notes: notes,
            //perms: perms, // TODO: validate on insert?
        }
        subspecies && (dataObj.subspecies = subspecies?._id)
        imageFile && formData.set("img", imageFile, "img")

        fetch(BaseExternalUrl + "/import/fruit", {
            method: 'Post',
            body: formData,
            headers: {
                credentials: 'include',
                'Cookie': cookies,
                // 'Content-type': "multipart/form-data" // TODO: auth?
                //Authorization: tokenFetch,
            },
        })
            .then(HandleTxtResponse) // TODO: make sure imports do it this way
            .then((newid) => {
                redirect(BaseExternalUrl + "/view/fruit/" + newid)
            })
            .catch((err) => {
                setErr(JSON.stringify(err))
            });
    }
    return <ImportEntryFormWrapper entryType={"fruit"}>
        <ErrorDisplay err={err} headerLevel={headerLevel}/>
        {/* Required Fields */}
        {/* TODO: ParentType: FOR "store" OR "outside" ONLY!!!!! */}{/* TODO: THIS!*/}
        <ExistingSpeciesSelector doSelect={setSpecies} headerLevel={headerLevel/*cookies={cookies}*/}/>
        {/* Optional fields*/}
        <ExistingSubSpeciesSelector species={species?._id} doSelect={setSubspecies}
                                    headerLevel={headerLevel/*cookies={cookies}*/}/>
        <ImageSelector updateParent={setImageFile}/>
        {/*<EntryPermsArea setEntryPerms={setPerms}/>*/}
        <NewEntryNotes setNotes={setNotes}/>
        {/* SUBMIT AREA */}
        <input type="submit" value="Submit" onClick={submitImportFruit} onSubmit={(e) => {
            e.preventDefault();
        }}/>
    </ImportEntryFormWrapper>
}

export function FruitInline({data, expandByDefault, onClick, showMainPageButton, idIsLink}: InlineProps<FruitData>) {
    // TODO: do these need depth providers? probably not
    const [expanded, setExpanded] = useState(expandByDefault)
    return <InlineEntry onClick={onClick}>
        <InlineSubArea props={{}}>
            <ID id={data._id} txt={"Fruit"} entryType={"fruit"} allowOpenMainPage={showMainPageButton}
                linkPage={idIsLink}/>
            <MostRecentImageDisplay data={data.mostRecentImage}/>
            <SpeciesArea readonly={true} initial={data.species}/>
            <SubspeciesArea readonly={true} currentSpecies={data.species} initialSub={data.subspecies}/>
            <DisposedDisplay readonly={true} disposed={data.disposed}/>
        </InlineSubArea>
        <InlineExpansionArea props={{expanded: expanded}}>
            <ParentDisplay parent={data.parent} parentType={data.parentType}/>
            <GensInlineDisplay gensSinceSpore={data.genSpore} dontDisplayGensFruitOrSpore={true}
            />
            <div>{"SPORE PRINTS AREA"}</div>
            {/* TODO: Prints? */}{/* TODO: ?????? */}
            {/* TODO: PROJECTS? ON OTHERS TOO */}
            <NotesAreaInline notes={data.notes} offset={-1}/>
            <DateArea pre={"Last updated: "} readonly={true} when={data.lastUpdated}/>
        </InlineExpansionArea>
        <InlineExpansionButton data-cy-id="InlineSubAreaButton" setExpanded={setExpanded}
                               expanded={expanded}/>
    </InlineEntry>
}

export function CreateCloneArea( // TODO: this vs NewFruitForm
    {
        fruitId, headerLevel, onCloneCreated, readonly, cookies,
    }: {
        fruitId: string,
        headerLevel?: number,
        onCloneCreated: (f: FruitData) => void,
        readonly: boolean,
        cookies: string,
    }) {
    if (readonly) {
        return null
    }
    const [typeTo, setTypeTo] = useState("plate")
    const [idTo, setIdTo] = useState<string | undefined>()
    const [notes, setNotes] = useState<Note[]>([])
    const [err, setErr] = useState<string | undefined>()
    const handleCreate = () => {
        // TODO: inherit perms?
        fetch(BaseExternalUrl + "/create/clone", { // TODO: ensure ok
            method: "POST",
            headers: {
                credentials: 'include',
                'Cookie': cookies,
                'Content-type': "application/json"
            },
            body: JSON.stringify({
                idFrom: fruitId,
                typeFrom: "fruit",
                typeTo: typeTo,
                idTo: idTo,
                notes: notes,
            })
        }).then(HandleJsonResponse).then((newEntry) => {
            try {
                AssertFruit(newEntry)
                onCloneCreated(newEntry)
                // TODO: WE HAVE CREATED A NEW FRUIT!!!!! NOTIFY???
                // TODO ? redirect(BaseUrl+"/view/fruit/"+newId) // TODO: ok?
            } catch (er) {
                setErr("failed to decode response:")
            }
        }).catch((er) => {
            setErr(JSON.stringify(er))
        });
    }
    return <div>
        <ErrorDisplay err={err} headerLevel={headerLevel}/>
        <div>
            <div>{"Create Clone:"}</div>
            <div>
                <TestAndValidate todos={["no need for type?"]}>
                    <div>{"TYPE TO:"}</div>
                </TestAndValidate>
                <select className={"tailwindSelector"} value={typeTo} onSelect={e => {
                    setTypeTo(e.currentTarget.value)
                }} onChange={() => {
                }}>
                    {["plate", "jar", "slant"].map((opt, i) => {
                        return <option value={opt} key={i}>{opt}</option>
                    })}
                </select>
            </div>
            <div>
                <TestAndValidate
                    todos={["validate that this is working properly in typing as well as reading from rfid"]}>
                    <NameArea currentName={idTo} setName={setIdTo} headerTxt={"Select ID: "} readonly={false}
                              headerLevel={headerLevel}/>
                </TestAndValidate>
                <ReadRFIDButton handleTagRead={setIdTo}/>
            </div>
        </div>
        <NewEntryNotes setNotes={setNotes}/>
        <button className={"basicButton"} onClick={e => {
            e.preventDefault()
            handleCreate()
        }}>{"Submit new Clone"}</button>
    </div>
}

// export function FruitListDisplay({data, onClick}: SingleListProps<FruitData>) {
//     return <div>
//         {data.map((b, i) => {
//             return <FruitInline data={b} onClick={() => {
//                 onClick(b)
//             }} key={i}/>
//         })}
//     </div>
// }

// TODO: HEAVILY TEST!!!!
export function FruitRecentSelector({onSelect}: { onSelect: (selected?: FruitData) => void }) {
    return <RecentSelectorV2<FruitData> listUrlType={"fruits"} assertion={AssertFruit} singleConstructor={(val, i) => {
        return <FruitInline data={val} expandByDefault={false} onClick={onSelect}/>
    }}/>
}

export function FruitListPageTable({data, onClick}: ListPageItems<FruitData>) {
    const cols: ListTableColumn<FruitData>[] = [
        NewColumn("ID", (v)=>v._id),
        NewColumn("Harvest", (v)=>{
            return NumberToDateStr(v.creationDate)
        }),
        NewColumn("Species", v=>v.species ),
        NewColumn("Subspecies", (v)=>v.subspecies || ""),
    ]
    // TODO: expansion for everything else????
    return <ListPageTable cols={cols} data={data} onClick={onClick}/>
}